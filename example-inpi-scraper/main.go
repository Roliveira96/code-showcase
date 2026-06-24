package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: arquivo .env não encontrado.")
	}

	searchTerm := os.Getenv("SEARCH_TERM")
	searchYearStr := os.Getenv("SEARCH_YEAR")

	if searchTerm == "" || searchYearStr == "" {
		log.Fatal("As variáveis SEARCH_TERM e SEARCH_YEAR precisam estar definidas no .env")
	}

	searchYear, err := strconv.Atoi(searchYearStr)
	if err != nil {
		log.Fatalf("Ano inválido: %v", err)
	}

	startRPI, endRPI := calcularRPIsDoAno(searchYear)
	searchTerm = strings.ToLower(searchTerm)

	fmt.Printf("Iniciando busca ultrarrápida (SEM NAVEGADOR) pelo termo: '%s'\n", searchTerm)
	fmt.Printf("Ano do calendário: %d\n", searchYear)
	fmt.Printf("Intervalo RPI (Apenas Marcas identificadas via 'RM'): %d até %d\n", startRPI, endRPI)

	downloadsDir := "revistas_inpi"
	if _, err := os.Stat(downloadsDir); os.IsNotExist(err) {
		os.Mkdir(downloadsDir, 0755)
	}

	leadsDir := filepath.Join("leads", strconv.Itoa(searchYear))
	if err := os.MkdirAll(leadsDir, 0755); err != nil {
		log.Fatalf("Erro ao criar pasta leads: %v", err)
	}

	// Relatório de Busca Pessoal do Usuário
	termoSanitizado := strings.ReplaceAll(searchTerm, " ", "_")
	relatorioPath := fmt.Sprintf("relatorio_inpi_busca_%s.txt", termoSanitizado)
	relatorioFile, err := os.Create(relatorioPath)
	if err != nil {
		log.Fatalf("Erro ao criar o arquivo de relatório: %v", err)
	}
	defer relatorioFile.Close()

	relatorioFile.WriteString(fmt.Sprintf("=======================================================\n"))
	relatorioFile.WriteString(fmt.Sprintf("RELATÓRIO DE BUSCA INPI - AUTOMATIZADO\n"))
	relatorioFile.WriteString(fmt.Sprintf("Termo Procurado: %s\n", searchTerm))
	relatorioFile.WriteString(fmt.Sprintf("Ano Base: %d\n", searchYear))
	relatorioFile.WriteString(fmt.Sprintf("=======================================================\n\n"))

	totalRevistasVerificadas := 0
	revistasComEncontro := 0

	for i := startRPI; i <= endRPI; i++ {
		totalRevistasVerificadas++
		fileName := filepath.Join(downloadsDir, fmt.Sprintf("RM%d.xml", i))
		achouNaRPI := false

		if _, err := os.Stat(fileName); err == nil {
			fmt.Printf("\n--- Verificando RPI %d [CACHE LOCAL] ---\n", i)
			achouNaRPI = processarXML(fileName, searchTerm, i, relatorioFile, leadsDir)
		} else {
			url := fmt.Sprintf("https://revistas.inpi.gov.br/txt/RM%d.zip", i)
			fmt.Printf("\n--- Verificando RPI %d [WEB] ---\n", i)
			fmt.Printf("Baixando e extraindo silenciosamente: %s\n", url)

			err := baixarEExtrairRevista(url, fileName)
			if err != nil {
				if strings.Contains(err.Error(), "404") {
					fmt.Println("Revista inexistente ou ainda não publicada (404).")
				} else {
					fmt.Printf("Erro no rpi %d: %v\n", i, err)
				}
				continue
			} else {
				achouNaRPI = processarXML(fileName, searchTerm, i, relatorioFile, leadsDir)
			}
		}

		if achouNaRPI {
			revistasComEncontro++
		}
	}

	relatorioFile.WriteString(fmt.Sprintf("\n=======================================================\n"))
	relatorioFile.WriteString(fmt.Sprintf("RESUMO FINAL DA BUSCA:\n"))
	relatorioFile.WriteString(fmt.Sprintf("Total de revistas vasculhadas: %d\n", totalRevistasVerificadas))
	relatorioFile.WriteString(fmt.Sprintf("Total de revistas onde a marca/processo apareceu: %d\n", revistasComEncontro))
	relatorioFile.WriteString(fmt.Sprintf("=======================================================\n"))

	fmt.Println("\n\n================ RESUMO FINAL ================")
	fmt.Printf("✔ Total de revistas vasculhadas: %d\n", totalRevistasVerificadas)
	fmt.Printf("✔ Total de revistas onde o termo '%s' apareceu: %d\n", searchTerm, revistasComEncontro)
	fmt.Printf("✔ Relatório detalhado da sua busca salvo em: 👉 %s\n", relatorioPath)
	fmt.Printf("🔥 CSVs de Leads divididos automaticamente na pasta 👉 /leads/\n")
	fmt.Println("==============================================")
}

func calcularRPIsDoAno(targetYear int) (int, int) {
	baselineDate := time.Date(2026, time.March, 31, 0, 0, 0, 0, time.UTC)
	baselineRPI := 2882

	startOfYear := time.Date(targetYear, time.January, 1, 0, 0, 0, 0, time.UTC)
	for startOfYear.Weekday() != time.Tuesday {
		startOfYear = startOfYear.AddDate(0, 0, 1)
	}
	diffStartWeeks := int(startOfYear.Sub(baselineDate).Hours() / 24 / 7)
	startRPI := baselineRPI + diffStartWeeks

	var endDate time.Time
	now := time.Now()
	// Se for o ano corrente, não tem como prever o futuro
	if targetYear == now.Year() {
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		// Se for ano passado, pega até a última semana do ano passado!
		endDate = time.Date(targetYear, time.December, 31, 0, 0, 0, 0, time.UTC)
	}

	for endDate.Weekday() != time.Tuesday {
		endDate = endDate.AddDate(0, 0, -1)
	}
	diffEndWeeks := int(endDate.Sub(baselineDate).Hours() / 24 / 7)
	endRPI := baselineRPI + diffEndWeeks

	if endRPI < startRPI {
		endRPI = startRPI
	}

	return startRPI, endRPI
}

func baixarEExtrairRevista(url string, saveFileName string) error {
	client := http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return fmt.Errorf("404")
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("status inesperado: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(bodyBytes), int64(len(bodyBytes)))
	if err != nil {
		return fmt.Errorf("arquivo zip corrompido ou invalido: %v", err)
	}

	for _, zipFile := range zipReader.File {
		if strings.HasSuffix(strings.ToLower(zipFile.Name), ".xml") {
			xmlFile, err := zipFile.Open()
			if err != nil {
				return err
			}

			out, err := os.Create(saveFileName)
			if err != nil {
				xmlFile.Close()
				return err
			}

			_, err = io.Copy(out, xmlFile)
			out.Close()
			xmlFile.Close()

			if err != nil {
				return err
			}
			fmt.Println("Tudo certo! Revista XML descompactada e arquivada localmente.")
			return nil
		}
	}

	return fmt.Errorf("nenhum XML de marca encontrado dentro do ZIP")
}

func extrairAtributo(linha string, param string) string {
	idx := strings.Index(linha, param)
	if idx == -1 {
		return ""
	}
	start := idx + len(param) + 1
	end := strings.Index(linha[start:], "\"")
	if end == -1 {
		return ""
	}
	conteudo := linha[start : start+end]
	conteudo = strings.ReplaceAll(conteudo, ";", ",")
	conteudo = strings.ReplaceAll(conteudo, "\n", " ")
	return conteudo
}

func extrairTag(linha string, tagInicio string, tagFim string) string {
	idxInicio := strings.Index(linha, tagInicio)
	if idxInicio == -1 {
		return ""
	}
	idxFim := strings.Index(linha[idxInicio:], tagFim)
	if idxFim == -1 {
		return ""
	}
	start := idxInicio + len(tagInicio)
	conteudo := linha[start : idxInicio+idxFim]
	conteudo = strings.ReplaceAll(conteudo, ";", ",")
	return conteudo
}

func processarXML(fileName, term string, numeroRPI int, logger *os.File, leadsDir string) bool {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Erro ao abrir cache XML: %v\n", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, 1024*1024)

	var blockBuilder strings.Builder
	found := false
	insideProcess := false
	qualquerAparicaoNaRevista := false

	// Configuração do Fatiamento dos CSVs (Máximo 2000 leads)
	var leadsFile *os.File
	leadsCounter := 0
	partesGerais := 1

	abrirNovoArquivoCSV := func() {
		if leadsFile != nil {
			leadsFile.Close()
		}
		pathCSV := filepath.Join(leadsDir, fmt.Sprintf("rv_%d_part_%d.csv", numeroRPI, partesGerais))
		leadsFile, _ = os.Create(pathCSV)
		leadsFile.WriteString("Revista;Processo;Razão Social;Estado;O que a empresa faz (Especificação);Status (Despacho)\n")
		partesGerais++
	}

	abrirNovoArquivoCSV() // Abre a "part 1" inicial
	
	// Previne arquivos não fechados caso o scanner saia por erro
	defer func() {
		if leadsFile != nil {
			leadsFile.Close()
		}
	}()

	var leadProcesso, leadRazao, leadUF, leadAtividade, leadDespacho string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "<processo ") {
			insideProcess = true
			blockBuilder.Reset()
			
			leadProcesso = extrairAtributo(line, "numero=")
			leadRazao = ""
			leadUF = ""
			leadAtividade = ""
			leadDespacho = ""
		}

		if insideProcess {
			blockBuilder.WriteString(line + "\n")

			if strings.Contains(line, "<despacho ") && leadDespacho == "" {
				leadDespacho = extrairAtributo(line, "nome=")
			}

			if strings.Contains(line, "<titular ") {
				leadRazao = extrairAtributo(line, "nome-razao-social=")
				leadUF = extrairAtributo(line, "uf=")
			}

			if strings.Contains(line, "<especificacao>") {
				leadAtividade = extrairTag(line, "<especificacao>", "</especificacao>")
			}
			
			if strings.Contains(strings.ToLower(line), term) {
				found = true
				qualquerAparicaoNaRevista = true
			}

			if strings.Contains(line, "</processo>") {
				// 1. Gera e empacota o LEAd do seu novo banco
				if leadRazao != "" {
					if leadsCounter >= 2000 {
						abrirNovoArquivoCSV()
						leadsCounter = 0
					}

					linhaLead := fmt.Sprintf("%d;%s;%s;%s;%s;%s\n",
						numeroRPI, leadProcesso, leadRazao, leadUF, leadAtividade, leadDespacho)
					leadsFile.WriteString(linhaLead)
					leadsCounter++
				}

				// 2. Extrai relatório da sua busca em específico
				if found {
					txtDoProcesso := blockBuilder.String()
					fmt.Printf("\n★★★ LOCALIZADO NA RPI %d ★★★\n", numeroRPI)
					fmt.Println(txtDoProcesso)
					fmt.Println("--------------------------")

					logger.WriteString(fmt.Sprintf(">>>>>>>> LOG ENCONTRADO NA REVISTA (RPI): %d <<<<<<<<\n", numeroRPI))
					logger.WriteString(txtDoProcesso)
					logger.WriteString("\n-------------------------------------------------------\n\n")
				}
				insideProcess = false
				found = false 
			}
		}
	}

	if !qualquerAparicaoNaRevista {
		fmt.Println("Nenhum registro sobre o termo procurado nesta revista.")
	}

	return qualquerAparicaoNaRevista
}
