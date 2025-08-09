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
	NomeRemetente    string
	EmailRemetente   string
	EmailCopiaOculta string
	NomeCopiaOculta  string
	Assunto          string
	IdTemplate       int
	Tipo             int
}

func gerarJsonDeCsv(caminhoCsv string, prefixoSaida string, maxEnvios int, config EmailConfig) error {
	arquivoCsv, err := os.Open(caminhoCsv)
	if err != nil {
		return fmt.Errorf("erro ao abrir o arquivo CSV: %w", err)
	}
	defer arquivoCsv.Close()

	leitor := csv.NewReader(arquivoCsv)
	leitor.Comma = ','

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

		emailRepresentante := linha[1]
		nomeRepresentante := linha[2]
		emailUsuarioFinanceiro := linha[3]
		nomeUsuarioFinanceiro := linha[4]
		emailUsuarioCientifico := linha[5]
		nomeUsuarioCientifico := linha[6]

		copiaMapper := make(map[string]string)
		if emailUsuarioFinanceiro != "" && nomeUsuarioFinanceiro != "" {
			copiaMapper[emailUsuarioFinanceiro] = nomeUsuarioFinanceiro
		}
		if emailUsuarioCientifico != "" && nomeUsuarioCientifico != "" {
			copiaMapper[emailUsuarioCientifico] = nomeUsuarioCientifico
		}

		// Inicializa o mapa de cópia oculta
		copiaOcultaMapper := make(map[string]string)
		// Adiciona o destinatário de cópia oculta padrão da configuração
		if config.EmailCopiaOculta != "" && config.NomeCopiaOculta != "" {
			copiaOcultaMapper[config.EmailCopiaOculta] = config.NomeCopiaOculta
		}

		payload := EmailPayload{
			NomeRemetente:  config.NomeRemetente,
			EmailRemetente: config.EmailRemetente,
			DestinatariosMapper: map[string]string{
				emailRepresentante: nomeRepresentante,
			},
			DestinatariosCopiaMapper:       copiaMapper,
			DestinatariosCopiaOcultaMapper: copiaOcultaMapper, // Usa o mapa atualizado
			IdTemplate:                     config.IdTemplate,
			Tipo:                           config.Tipo,
			Assunto:                        config.Assunto,
			Variaveis: map[string]string{
				"NomeCliente": GetFirstName(nomeRepresentante),
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

	config := EmailConfig{
		NomeRemetente:    os.Getenv("NOME_REMETENTE"),
		EmailRemetente:   os.Getenv("EMAIL_REMETENTE"),
		EmailCopiaOculta: os.Getenv("EMAIL_COPIA_OCULTA"),
		NomeCopiaOculta:  os.Getenv("NOME_COPIA_OCULTA"),
		Assunto:          os.Getenv("ASSUNTO_EMAIL"),
		IdTemplate:       idTemplate,
		Tipo:             tipo,
	}

	caminhoCsv := "teste1.csv"
	prefixoSaida := "emails"
	maxEnviosPorArquivo := 10

	fmt.Printf("Gerando arquivos JSON a partir de '%s' com no máximo %d envios por arquivo...\n", caminhoCsv, maxEnviosPorArquivo)

	err = gerarJsonDeCsv(caminhoCsv, prefixoSaida, maxEnviosPorArquivo, config)
	if err != nil {
		fmt.Printf("Ocorreu um erro: %v\n", err)
		return
	}

	fmt.Println("Processo de geração de arquivos JSON concluído.")
}
