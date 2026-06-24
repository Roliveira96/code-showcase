package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type EmailPayload struct {
	NomeRemetente                  string            `json:"nomeRemetente"`
	EmailRemetente                 string            `json:"emailRemetente"`
	DestinatariosMapper            map[string]string `json:"destinatariosMapper"`
	DestinatariosCopiaMapper       map[string]string `json:"destinatariosCopiaMapper"`
	DestinatariosCopiaOcultaMapper map[string]string `json:"destinatariosCopiaOcultaMapper"`
	IdTemplate                     int               `json:"idTemplate"`
	Tipo                           int               `json:"tipo"`
	Assunto                        string            `json:"assunto"`
	Variaveis                      map[string]string `json:"variaveis"`
}

type EmailConfig struct {
	NomeRemetente     string
	EmailRemetente    string
	EmailsEmCopia     []string
	NomesEmCopia      []string
	EmailsCopiaOculta []string
	NomesCopiaOculta  []string
	Assunto           string
	IdTemplate        int
	Tipo              int
}

// decodeISO88591ToUTF8 converte bytes de ISO-8859-1 para UTF-8.
// ISO-8859-1 mapeia 1:1 para os primeiros 256 code points do Unicode.
func decodeISO88591ToUTF8(iso88591 []byte) []byte {
	var buf bytes.Buffer
	buf.Grow(len(iso88591))
	for _, b := range iso88591 {
		buf.WriteRune(rune(b))
	}
	return buf.Bytes()
}

// sanitizeString remove caracteres que podem quebrar templates ou causar problemas
func sanitizeString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "[", "")
	s = strings.ReplaceAll(s, "]", "")
	return s
}

func gerarJsonDeCsv(caminhoCsv string, prefixoSaida string, maxEnvios int, config EmailConfig) error {
	// Lê o arquivo inteiro
	conteudo, err := os.ReadFile(caminhoCsv)
	if err != nil {
		return fmt.Errorf("erro ao abrir o arquivo CSV: %w", err)
	}

	// Converte o conteúdo para UTF-8
	conteudoUtf8 := decodeISO88591ToUTF8(conteudo)

	// Cria o leitor CSV a partir do conteúdo convertido
	leitor := csv.NewReader(bytes.NewReader(conteudoUtf8))
	leitor.Comma = ';'
	leitor.LazyQuotes = true

	// Lê e descarta a linha do cabeçalho
	_, err = leitor.Read()
	if err != nil {
		return fmt.Errorf("erro ao ler o cabeçalho do CSV: %w", err)
	}

	var payloads []EmailPayload
	fileCount := 1
	payloadCount := 0

	for {
		linha, err := leitor.Read()
		if err == io.EOF {
			if len(payloads) > 0 {
				nomeArquivoSaida := fmt.Sprintf("%s_%d.json", prefixoSaida, fileCount)
				if err := writePayloadsToFile(payloads, nomeArquivoSaida); err != nil {
					return err
				}
				fmt.Printf("Arquivo JSON '%s' gerado com sucesso.\n", nomeArquivoSaida)
			}
			break
		}
		if err != nil {
			return fmt.Errorf("erro ao ler uma linha do CSV: %w", err)
		}

		// Ajuste para o novo formato do CSV:
		// linha[0] -> Razao Social
		// linha[1] -> Email Representante
		// linha[2] -> Nome Representante
		// linha[3] -> Email Financeiro
		// linha[4] -> Nome Financeiro
		// linha[5] -> Email Cientifico
		// linha[6] -> Nome Cientifico
		// linha[7] -> Email Adocao
		// linha[8] -> Nome Adocao
		if len(linha) < 3 {
			fmt.Printf("Linha ignorada (formato inválido): %v\n", linha)
			continue
		}
		nomeRepresentante := sanitizeString(linha[2])
		emailRepresentante := sanitizeString(linha[1])

		copiaMapper := make(map[string]string)

		// Adiciona destinatários em cópia (CC) do arquivo .env
		if len(config.EmailsEmCopia) > 0 && len(config.EmailsEmCopia) == len(config.NomesEmCopia) {
			for i, email := range config.EmailsEmCopia {
				email = sanitizeString(email)
				nome := sanitizeString(config.NomesEmCopia[i])
				if email != "" && nome != "" {
					copiaMapper[email] = nome
				}
			}
		}

		// Adiciona destinatários em cópia do CSV
		colunasCopia := []struct {
			emailIdx int
			nomeIdx  int
		}{
			{3, 4}, // Financeiro
			{5, 6}, // Cientifico
			{7, 8}, // Adocao
		}

		for _, col := range colunasCopia {
			if len(linha) > col.nomeIdx {
				email := sanitizeString(linha[col.emailIdx])
				nome := sanitizeString(linha[col.nomeIdx])
				if email != "" {
					if nome == "" {
						nome = email // Fallback se não tiver nome
					}
					copiaMapper[email] = nome
				}
			}
		}

		// Adiciona destinatários em cópia oculta (BCC) do arquivo .env
		copiaOcultaMapper := make(map[string]string)
		if len(config.EmailsCopiaOculta) > 0 && len(config.EmailsCopiaOculta) == len(config.NomesCopiaOculta) {
			for i, email := range config.EmailsCopiaOculta {
				email = sanitizeString(email)
				nome := sanitizeString(config.NomesCopiaOculta[i])
				if email != "" && nome != "" {
					copiaOcultaMapper[email] = nome
				}
			}
		}

		payload := EmailPayload{
			NomeRemetente:  config.NomeRemetente,
			EmailRemetente: config.EmailRemetente,
			DestinatariosMapper: map[string]string{
				emailRepresentante: nomeRepresentante,
			},
			DestinatariosCopiaMapper:       copiaMapper,
			DestinatariosCopiaOcultaMapper: copiaOcultaMapper,
			IdTemplate:                     config.IdTemplate,
			Tipo:                           config.Tipo,
			Assunto:                        config.Assunto,
			Variaveis: map[string]string{
				"FirstName": GetFirstName(nomeRepresentante),
			},
		}

		payloads = append(payloads, payload)
		payloadCount++

		if payloadCount >= maxEnvios {
			nomeArquivoSaida := fmt.Sprintf("%s_%d.json", prefixoSaida, fileCount)
			if err := writePayloadsToFile(payloads, nomeArquivoSaida); err != nil {
				return err
			}
			fmt.Printf("Arquivo JSON '%s' gerado com sucesso.\n", nomeArquivoSaida)

			payloads = nil
			payloadCount = 0
			fileCount++
		}
	}

	return nil
}

func writePayloadsToFile(payloads []EmailPayload, caminhoTxt string) error {
	dadosJson, err := json.MarshalIndent(payloads, "", "    ")
	if err != nil {
		return fmt.Errorf("erro ao converter dados para JSON: %w", err)
	}

	err = os.WriteFile(caminhoTxt, dadosJson, 0644)
	if err != nil {
		return fmt.Errorf("erro ao escrever no arquivo de saída '%s': %w", caminhoTxt, err)
	}
	return nil
}

func GetFirstName(fullName string) string {
	names := strings.Split(fullName, " ")
	if len(names) > 0 {
		return names[0]
	}
	return ""
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Erro ao carregar o arquivo .env, usando valores padrão se disponíveis.")
	}

	idTemplate, _ := strconv.Atoi(os.Getenv("ID_TEMPLATE"))
	tipo, _ := strconv.Atoi(os.Getenv("TIPO_EMAIL"))

	// Carrega e-mails em cópia (CC)
	emailsEmCopiaStr := os.Getenv("EMAILS_EM_COPIA")
	nomesEmCopiaStr := os.Getenv("NOMES_EM_COPIA")
	var emailsEmCopia []string
	if emailsEmCopiaStr != "" {
		emailsEmCopia = strings.Split(emailsEmCopiaStr, ",")
	}
	var nomesEmCopia []string
	if nomesEmCopiaStr != "" {
		nomesEmCopia = strings.Split(nomesEmCopiaStr, ",")
	}

	// Carrega e-mails em cópia oculta (BCC)
	emailsCopiaOcultaStr := os.Getenv("EMAILS_COPIA_OCULTA")
	nomesCopiaOcultaStr := os.Getenv("NOMES_COPIA_OCULTA")
	var emailsCopiaOculta []string
	if emailsCopiaOcultaStr != "" {
		emailsCopiaOculta = strings.Split(emailsCopiaOcultaStr, ",")
	}
	var nomesCopiaOculta []string
	if nomesCopiaOcultaStr != "" {
		nomesCopiaOculta = strings.Split(nomesCopiaOcultaStr, ",")
	}

	config := EmailConfig{
		NomeRemetente:     os.Getenv("NOME_REMETENTE"),
		EmailRemetente:    os.Getenv("EMAIL_REMETENTE"),
		EmailsEmCopia:     emailsEmCopia,
		NomesEmCopia:      nomesEmCopia,
		EmailsCopiaOculta: emailsCopiaOculta,
		NomesCopiaOculta:  nomesCopiaOculta,
		Assunto:           os.Getenv("ASSUNTO_EMAIL"),
		IdTemplate:        idTemplate,
		Tipo:              tipo,
	}

	caminhoCsv := "49_DisparoFomentoDireto.CSV" // Lembre-se de usar o nome do seu novo arquivo CSV
	prefixoSaida := "emails"
	maxEnviosPorArquivo := 2000

	fmt.Printf("Gerando arquivos JSON a partir de '%s' com no máximo %d envios por arquivo...\n", caminhoCsv, maxEnviosPorArquivo)

	err = gerarJsonDeCsv(caminhoCsv, prefixoSaida, maxEnviosPorArquivo, config)
	if err != nil {
		fmt.Printf("Ocorreu um erro: %v\n", err)
		return
	}

	fmt.Println("Processo de geração de arquivos JSON concluído.")
}
