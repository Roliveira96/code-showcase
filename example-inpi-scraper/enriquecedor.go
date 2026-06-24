//go:build ignore
// +build ignore

package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type BrasilAPIResponse struct {
	Cnpj        string `json:"cnpj"`
	RazaoSocial string `json:"razao_social"`
	Email       string `json:"email"`
	DddTelef1   string `json:"ddd_telefone_1"`
	Telefone1   string `json:"telefone_1"`
}

func main() {
	godotenv.Load()
	ano := os.Getenv("SEARCH_YEAR")
	if ano == "" {
		ano = "2026"
	}
	baseDir := filepath.Join("leads", ano)

	fmt.Printf("=== ENRIQUECEDOR DE LEADS INPI (%s) ===\n", ano)
	fmt.Printf("Varrendo a pasta: %s\n\n", baseDir)

	files, err := os.ReadDir(baseDir)
	if err != nil {
		log.Fatal("Não encontrei a pasta de leads para este ano: ", baseDir)
	}

	// Arquivo MESTRE para consolidar todos os leads "RICOS" (com contato de fato)
	masterPath := "leads_ricos_consolidados.csv"
	masterFile, _ := os.Create(masterPath)
	defer masterFile.Close()
	masterWriter := csv.NewWriter(masterFile)
	masterWriter.Comma = ';'
	
	headerEscrito := false

	for _, file := range files {
		// Pula arquivos que já foram enriquecidos
		if !strings.HasSuffix(file.Name(), ".csv") || strings.Contains(file.Name(), "_RICO") {
			continue
		}

		filePath := filepath.Join(baseDir, file.Name())
		processarCSV(filePath, masterWriter, &headerEscrito)
	}
	
	fmt.Printf("\n=== PROCESSADO! Arquivo mestre gerado: %s ===\n", masterPath)
}

func processarCSV(filePath string, masterWriter *csv.Writer, headerEscrito *bool) {
	fmt.Println("------------- INICIANDO ------------")
	fmt.Println("Lendo Arquivo:", filePath)

	fOriginal, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Erro ao ler:", err)
		return
	}
	defer fOriginal.Close()

	reader := csv.NewReader(fOriginal)
	reader.Comma = ';'
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Erro CSV:", err)
		return
	}

	if len(records) < 2 {
		return
	}

	novaRota := strings.Replace(filePath, ".csv", "_RICO.csv", 1)
	fNovo, _ := os.Create(novaRota)
	defer fNovo.Close()
	writer := csv.NewWriter(fNovo)
	writer.Comma = ';'

	// Atualizamos o cabeçalho original com nossas 4 novas colunas Douradas
	header := records[0]
	header = append(header, "CNPJ", "Email", "Telefone", "Domínio", "URL de Busca")
	writer.Write(header)

	// Escreve o cabeçalho no arquivo mestre apenas uma vez
	if !*headerEscrito {
		masterWriter.Write(header)
		masterWriter.Flush()
		*headerEscrito = true
	}

	totalRegistros := len(records) - 1
	fmt.Printf("=> O arquivo tem %d empresas para pesquisar.\n", totalRegistros)
	
	for i, record := range records[1:] {
		razao := record[2]
		uf := record[3]

		// Limpa caracteres especiais que as vezes vem no XML do INPI
		razaoLimpa := strings.ReplaceAll(razao, "(", "")
		razaoLimpa = strings.ReplaceAll(razaoLimpa, ")", "")
		
		fmt.Printf("[%d/%d] Buscando: %s... ", i+1, totalRegistros, razaoLimpa)

		urlBusca := ""
		cnpj := ""
		// Tenta extrair o CNPJ raiz de 8 dígitos se ele já estiver no nome (Muito comum no INPI)
		reRaiz := regexp.MustCompile(`\d{2}\.?\d{3}\.?\d{3}`)
		matchRaiz := reRaiz.FindString(razaoLimpa)
		if matchRaiz != "" {
			matchRaiz = regexp.MustCompile(`[^\d]`).ReplaceAllString(matchRaiz, "")
			if len(matchRaiz) == 8 {
				fmt.Printf("Usando Raiz do Nome... ")
				cnpj = matchRaiz + "0001" // Tenta supor matriz
				urlBusca = "Raiz extraída do nome no INPI"
			}
		}

		if cnpj == "" {
			cnpj, urlBusca = cacadorDeCNPJ(razaoLimpa, uf)
		}

		if cnpj != "" {
			// Se o CNPJ achado for apenas 8 ou 12 digitos, tenta completar para o teste
			if len(cnpj) == 8 {
			    cnpj = cnpj + "0001"
			}
			
			fmt.Printf("CNPJ Achado (%s). Puxando Dados da Receita... ", cnpj)
			email, tel := puxarDadosBrasilAPI(cnpj)
			
			dominio := ""
			email = strings.ToLower(email)
			
			// Se o email não for gratuito, extrai o dominio para abordar a empresa
			if email != "" && !strings.Contains(email, "@gmail.") && !strings.Contains(email, "@hotmail.") && !strings.Contains(email, "@yahoo.") && !strings.Contains(email, "@outlook.") {
			    parts := strings.Split(email, "@")
			    if len(parts) == 2 {
			        dominio = "www." + parts[1]
			    }
			}

			if email == "" || tel == "" {
				fmt.Printf("Aumentando precisão (CNPJ.WS)... ")
				emailSec, telSec := puxarDadosCNPJWS(cnpj)
				if email == "" {
				    email = emailSec
				}
				tel = telSec // Tenta o telefone secundário sempre pois formatação é melhor
			}

			if email != "" || tel != "" {
				fmt.Printf("=> CONTATO ENCONTRADO!\n")
				record = append(record, cnpj, email, tel, dominio, urlBusca)
				
				// Salva no arquivo individual _RICO
				writer.Write(record)
				writer.Flush() 
				
				// Salva no arquivo MESTRE consolidado
				masterWriter.Write(record)
				masterWriter.Flush()
			} else {
				fmt.Printf("=> OK (Sem contato público - Pulado).\n")
			}
		} else {
			fmt.Printf("Inconclusivo na Web.\n")
		}
		
		// LENTIDÃO ESTRATÉGICA ALEATÓRIA
		// Pausamos o script para evitar bloqueios das APIs e Buscadores
		// Usamos um valor aleatório entre 4 e 7 segundos para parecer mais humano
		atraso := 4 + (time.Now().UnixNano() % 4)
		time.Sleep(time.Duration(atraso) * time.Second) 
	}
	fmt.Println("Finalizado! Planilha com Leads Reais salva em:", novaRota)
}

func cacadorDeCNPJ(razao, uf string) (string, string) {
	termo := fmt.Sprintf("CNPJ \"%s\" %s", razao, uf)
	
	// Tentamos primeiro o DuckDuckGo LITE (Mais veloz e menos bloqueios que a versão HTML full)
	buscaDDG := "https://duckduckgo.com/lite/?q=" + url.QueryEscape(termo)
	fmt.Printf("(Busca DDG LITE: %s)... ", buscaDDG)
	cnpj := buscarNaWeb(buscaDDG)
	
	if cnpj == "" {
	    // Caso o DDG falhe ou bloqueie, tentamos o Yahoo como Plano B
	    buscaYahoo := "https://search.yahoo.com/search?p=" + url.QueryEscape(termo)
	    fmt.Printf("\n(Plano B - Yahoo: %s)... ", buscaYahoo)
	    cnpj = buscarNaWeb(buscaYahoo)
	    return cnpj, buscaYahoo
	}
	
	return cnpj, buscaDDG
}

func buscarNaWeb(urlBusca string) string {
	req, _ := http.NewRequest("GET", urlBusca, nil)
	// User-Agents aleatórios
	uas := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0",
		"Mozilla/5.0 (X11; Linux x86_64) Firefox/120.0",
	}
	req.Header.Set("User-Agent", uas[time.Now().Unix()%2])

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	html := string(bodyBytes)

	re := regexp.MustCompile(`\d{2}\.\d{3}\.\d{3}/\d{4}-\d{2}`)
	match := re.FindString(html)
	if match == "" {
		reClean := regexp.MustCompile(`\d{14}`)
		match = reClean.FindString(html)
	}

	if match != "" {
		match = regexp.MustCompile(`[^\d]`).ReplaceAllString(match, "") 
	}
	return match
}

func puxarDadosBrasilAPI(cnpj string) (string, string) {
	reqUrl := "https://brasilapi.com.br/api/cnpj/v1/" + cnpj
	
	req, _ := http.NewRequest("GET", reqUrl, nil)
	// Brasil API pede qualquer user-agent diferente do curl padrão
	req.Header.Set("User-Agent", "Bot-Pesquisa-Empresarial-INPI")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	
	if err != nil || resp.StatusCode != 200 {
		return "", ""
	}
	defer resp.Body.Close()

	var apiResp BrasilAPIResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)

	telefone := ""
	if apiResp.DddTelef1 != "" || apiResp.Telefone1 != "" {
		telefone = fmt.Sprintf("(%s) %s", apiResp.DddTelef1, apiResp.Telefone1)
	}

	return apiResp.Email, telefone
}

func puxarDadosCNPJWS(cnpj string) (string, string) {
	reqUrl := "https://publica.cnpj.ws/cnpj/" + cnpj
	
	req, _ := http.NewRequest("GET", reqUrl, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	
	if err != nil || resp.StatusCode != 200 {
		return "", ""
	}
	defer resp.Body.Close()

	var data struct {
		Estabelecimento struct {
			Email     string `json:"email"`
			Ddd1      string `json:"ddd1"`
			Telefone1 string `json:"telefone1"`
		} `json:"estabelecimento"`
	}
	json.NewDecoder(resp.Body).Decode(&data)

	tel := ""
	if data.Estabelecimento.Ddd1 != "" {
	    tel = fmt.Sprintf("(%s) %s", data.Estabelecimento.Ddd1, data.Estabelecimento.Telefone1)
	}

	return data.Estabelecimento.Email, tel
}
