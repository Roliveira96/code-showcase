package main

import (
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

func gerarJsonDeCsv(caminhoCsv string, prefixoSaida string, maxEnvios int, config EmailConfig) error {
	arquivoCsv, err := os.Open(caminhoCsv)
	if err != nil {
		return fmt.Errorf("erro ao abrir o arquivo CSV: %w", err)
	}
	defer arquivoCsv.Close()

	leitor := csv.NewReader(arquivoCsv)
	leitor.Comma = ','

	// Lê e descarta a linha do cabeçalho ("Nome", "Email")
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
		// linha[0] -> Nome
		// linha[1] -> Email
		nomeRepresentante := linha[0]
		emailRepresentante := linha[1]

		copiaMapper := make(map[string]string)

		// Adiciona destinatários em cópia (CC) do arquivo .env
		if len(config.EmailsEmCopia) > 0 && len(config.EmailsEmCopia) == len(config.NomesEmCopia) {
			for i, email := range config.EmailsEmCopia {
				email = strings.TrimSpace(email)
				nome := strings.TrimSpace(config.NomesEmCopia[i])
				if email != "" && nome != "" {
					copiaMapper[email] = nome
				}
			}
		} else if len(config.EmailsEmCopia) != len(config.NomesEmCopia) {
			fmt.Println("Aviso: Disparidade no número de e-mails e nomes em cópia no arquivo .env. Ignorando cópias do .env.")
		}

		// Adiciona destinatários em cópia oculta (BCC) do arquivo .env
		copiaOcultaMapper := make(map[string]string)
		if len(config.EmailsCopiaOculta) > 0 && len(config.EmailsCopiaOculta) == len(config.NomesCopiaOculta) {
			for i, email := range config.EmailsCopiaOculta {
				email = strings.TrimSpace(email)
				nome := strings.TrimSpace(config.NomesCopiaOculta[i])
				if email != "" && nome != "" {
					copiaOcultaMapper[email] = nome
				}
			}
		} else if len(config.EmailsCopiaOculta) != len(config.NomesCopiaOculta) {
			fmt.Println("Aviso: Disparidade no número de e-mails e nomes em cópia oculta no arquivo .env. Ignorando cópias ocultas.")
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

	caminhoCsv := "teste1.csv" // Lembre-se de usar o nome do seu novo arquivo CSV
	prefixoSaida := "emails"
	maxEnviosPorArquivo := 20

	fmt.Printf("Gerando arquivos JSON a partir de '%s' com no máximo %d envios por arquivo...\n", caminhoCsv, maxEnviosPorArquivo)

	err = gerarJsonDeCsv(caminhoCsv, prefixoSaida, maxEnviosPorArquivo, config)
	if err != nil {
		fmt.Printf("Ocorreu um erro: %v\n", err)
		return
	}

	fmt.Println("Processo de geração de arquivos JSON concluído.")
}
