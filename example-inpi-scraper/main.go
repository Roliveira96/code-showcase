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

	// Criação do Relatório em Texto
	termoSanitizado := strings.ReplaceAll(searchTerm, " ", "_")
	relatorioPath := fmt.Sprintf("relatorio_inpi_busca_%s.txt", termoSanitizado)
	relatorioFile, err := os.Create(relatorioPath)
	if err != nil {
		log.Fatalf("Erro ao criar o arquivo de relatório: %v", err)
	}
	defer relatorioFile.Close()

	// Cabeçalho do relatório
	relatorioFile.WriteString(fmt.Sprintf("=======================================================\n"))
	relatorioFile.WriteString(fmt.Sprintf("RELATÓRIO DE BUSCA INPI - AUTOMATIZADO\n"))
	relatorioFile.WriteString(fmt.Sprintf("Termo Procurado: %s\n", searchTerm))
	relatorioFile.WriteString(fmt.Sprintf("Ano Base: %d\n", searchYear))
	relatorioFile.WriteString(fmt.Sprintf("Intervalo de Revistas (RPIs): %d até %d\n", startRPI, endRPI))
	relatorioFile.WriteString(fmt.Sprintf("Data da Busca: %s\n", time.Now().Format("02/01/2006 15:04:05")))
	relatorioFile.WriteString(fmt.Sprintf("=======================================================\n\n"))

	// Variáveis para o Resumo Final
	totalRevistasVerificadas := 0
	revistasComEncontro := 0

	for i := startRPI; i <= endRPI; i++ {
		totalRevistasVerificadas++
		fileName := filepath.Join(downloadsDir, fmt.Sprintf("RM%d.xml", i))
		achouNaRPI := false

		if _, err := os.Stat(fileName); err == nil {
			fmt.Printf("\n--- Verificando RPI %d [CACHE LOCAL] ---\n", i)
			achouNaRPI = processarXML(fileName, searchTerm, i, relatorioFile)
		} else {
			// Download Direto se não tem cache
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
				// Se der erro/404, não tem como analisar
			} else {
				achouNaRPI = processarXML(fileName, searchTerm, i, relatorioFile)
			}
		}

		if achouNaRPI {
			revistasComEncontro++
		}
	}

	// Gravando resumo final no TXT
	relatorioFile.WriteString(fmt.Sprintf("\n=======================================================\n"))
	relatorioFile.WriteString(fmt.Sprintf("RESUMO FINAL DA BUSCA:\n"))
	relatorioFile.WriteString(fmt.Sprintf("Total de revistas vasculhadas: %d\n", totalRevistasVerificadas))
	relatorioFile.WriteString(fmt.Sprintf("Total de revistas onde a marca/processo apareceu: %d\n", revistasComEncontro))
	relatorioFile.WriteString(fmt.Sprintf("=======================================================\n"))

	// Resumo final impresso na tela
	fmt.Println("\n\n================ RESUMO FINAL ================")
	fmt.Printf("✔ Total de revistas vasculhadas: %d\n", totalRevistasVerificadas)
	fmt.Printf("✔ Total de revistas onde o termo '%s' apareceu: %d\n", searchTerm, revistasComEncontro)
	fmt.Printf("✔ Relatório detalhado salvo no seu projeto em:\n 👉 %s\n", relatorioPath)
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

	now := time.Now()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for currentDate.Weekday() != time.Tuesday {
		currentDate = currentDate.AddDate(0, 0, -1)
	}
	diffNowWeeks := int(currentDate.Sub(baselineDate).Hours() / 24 / 7)
	endRPI := baselineRPI + diffNowWeeks

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

// Analisa e agora retorna um booleano (se achou nessa revista ou não) e salva no relatório
func processarXML(fileName, term string, numeroRPI int, logger *os.File) bool {
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

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "<processo ") {
			insideProcess = true
			blockBuilder.Reset()
		}

		if insideProcess {
			blockBuilder.WriteString(line + "\n")

			if strings.Contains(strings.ToLower(line), term) {
				found = true
				qualquerAparicaoNaRevista = true
			}

			if strings.Contains(line, "</processo>") {
				if found {
					txtDoProcesso := blockBuilder.String()
					// Imprime no console pra você ver rodando
					fmt.Printf("\n★★★ LOCALIZADO NA RPI %d ★★★\n", numeroRPI)
					fmt.Println(txtDoProcesso)
					fmt.Println("--------------------------")

					// E também imprime no Relatório Físico (.txt)!
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
