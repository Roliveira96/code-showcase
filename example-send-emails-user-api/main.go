package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type EmailPayload struct {
	NomeRemetente                  string            `json:"nomeRemetente"`
	EmailRemetente                 string            `json:"emailRemetente"`
	DestinatariosMapper            map[string]string `json:"destinatariosMapper"`
	DestinatariosCopiaMapper       map[string]string `json:"destinatariosCopiaMapper,omitempty"`
	DestinatariosCopiaOcultaMapper map[string]string `json:"destinatariosCopiaOcultaMapper,omitempty"`
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
	ApiUrl            string
	AuthToken         string
}

func enviarEmails(payloads []EmailPayload, config EmailConfig) error {
	dadosJson, err := json.Marshal(payloads)
	if err != nil {
		return fmt.Errorf("erro ao converter payload para JSON: %w", err)
	}

	req, err := http.NewRequest("POST", config.ApiUrl, bytes.NewBuffer(dadosJson))
	if err != nil {
		return fmt.Errorf("erro ao criar a requisição HTTP: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.AuthToken)

	client := &http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao enviar a requisição HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("Lote de %d emails enviado com sucesso! Status: %s\n", len(payloads), resp.Status)
	} else {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("falha no envio do lote de emails. Status: %s, Resposta: %s", resp.Status, string(body))
	}

	return nil
}

func processarCsvEEnviarEmails(caminhoCsv string, maxEnviosPorLote int, config EmailConfig) error {
	arquivoCsv, err := os.Open(caminhoCsv)
	if err != nil {
		return fmt.Errorf("erro ao abrir o arquivo CSV: %w", err)
	}
	defer arquivoCsv.Close()

	leitor := csv.NewReader(arquivoCsv)
	leitor.Comma = ','

	// Lê e descarta a linha do cabeçalho
	_, err = leitor.Read()
	if err != nil {
		return fmt.Errorf("erro ao ler o cabeçalho do CSV: %w", err)
	}

	var payloads []EmailPayload
	loteCount := 1
	payloadCount := 0
	errosEnvio := 0

	for {
		linha, err := leitor.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("erro ao ler uma linha do CSV: %w", err)
		}

		if len(linha) < 3 {
			fmt.Printf("Aviso: pulando linha mal formatada no CSV: %v\n", linha)
			continue
		}

		primeiroNome := linha[0]
		sobrenome := linha[1]
		emailRepresentante := linha[2]
		nomeCompleto := strings.TrimSpace(primeiroNome + " " + sobrenome)

		copiaMapper := make(map[string]string)
		if len(config.EmailsEmCopia) > 0 && len(config.EmailsEmCopia) == len(config.NomesEmCopia) {
			for i, email := range config.EmailsEmCopia {
				email = strings.TrimSpace(email)
				nome := strings.TrimSpace(config.NomesEmCopia[i])
				if email != "" && nome != "" {
					copiaMapper[email] = nome
				}
			}
		} else if len(config.EmailsEmCopia) > 0 && config.EmailsEmCopia[0] != "" {
			fmt.Println("Aviso: Disparidade no número de e-mails e nomes em cópia no .env. Ignorando cópias.")
		}

		copiaOcultaMapper := make(map[string]string)
		if len(config.EmailsCopiaOculta) > 0 && len(config.EmailsCopiaOculta) == len(config.NomesCopiaOculta) {
			for i, email := range config.EmailsCopiaOculta {
				email = strings.TrimSpace(email)
				nome := strings.TrimSpace(config.NomesCopiaOculta[i])
				if email != "" && nome != "" {
					copiaOcultaMapper[email] = nome
				}
			}
		} else if len(config.EmailsCopiaOculta) > 0 && config.EmailsCopiaOculta[0] != "" {
			fmt.Println("Aviso: Disparidade no número de e-mails e nomes em cópia oculta no .env. Ignorando cópias ocultas.")
		}

		payload := EmailPayload{
			NomeRemetente:                  config.NomeRemetente,
			EmailRemetente:                 config.EmailRemetente,
			DestinatariosMapper:            map[string]string{emailRepresentante: nomeCompleto},
			DestinatariosCopiaMapper:       copiaMapper,
			DestinatariosCopiaOcultaMapper: copiaOcultaMapper,
			IdTemplate:                     config.IdTemplate,
			Tipo:                           config.Tipo,
			Assunto:                        config.Assunto,
			Variaveis:                      map[string]string{"FIRST_NAME": primeiroNome},
		}

		payloads = append(payloads, payload)
		payloadCount++

		if payloadCount >= maxEnviosPorLote {
			fmt.Printf("Enviando lote %d com %d emails...\n", loteCount, len(payloads))
			if err := enviarEmails(payloads, config); err != nil {
				fmt.Printf("Erro ao enviar o lote %d: %v\n", loteCount, err)
				errosEnvio++
			}
			payloads = nil
			payloadCount = 0
			loteCount++
			time.Sleep(1 * time.Second)
		}
	}

	if len(payloads) > 0 {
		fmt.Printf("Enviando lote final %d com %d emails...\n", loteCount, len(payloads))
		if err := enviarEmails(payloads, config); err != nil {
			fmt.Printf("Erro ao enviar o lote final %d: %v\n", loteCount, err)
			errosEnvio++
		}
	}

	if errosEnvio > 0 {
		return fmt.Errorf("%d lotes falharam ao enviar", errosEnvio)
	}

	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Aviso: Não foi possível carregar o arquivo .env. O programa tentará usar as variáveis de ambiente do sistema.")
	}

	idTemplate, _ := strconv.Atoi(os.Getenv("ID_TEMPLATE"))
	tipo, _ := strconv.Atoi(os.Getenv("TIPO_EMAIL"))

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
		ApiUrl:            os.Getenv("API_URL"),
		AuthToken:         os.Getenv("AUTH_TOKEN"),
	}

	if config.ApiUrl == "" || config.AuthToken == "" {
		fmt.Println("Erro: As variáveis API_URL e AUTH_TOKEN devem estar definidas no arquivo .env ou no ambiente do sistema.")
		return
	}

	caminhoCsv := "envio.csv"
	maxEnviosPorLote := 20

	fmt.Printf("Iniciando envio de emails a partir de '%s' com no máximo %d envios por lote...\n", caminhoCsv, maxEnviosPorLote)

	err = processarCsvEEnviarEmails(caminhoCsv, maxEnviosPorLote, config)
	if err != nil {
		fmt.Printf("\nOcorreu um erro durante o processo: %v\n", err)
		return
	}

	fmt.Println("\nProcesso de envio de emails concluído.")
}
