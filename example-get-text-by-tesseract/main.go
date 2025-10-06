package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	document "baliance.com/gooxml/document"
)

// Constantes de configuração
const (
	tikaServerURL = "http://localhost:9998/tika"
	pdfsDir       = "pdfs"
	resultsDir    = "results"
	trainingDir   = "training"
)

// trainingData define o nome do arquivo e a URL para download.
type trainingData struct {
	fileName string
	url      string
}

// Lista de arquivos de treinamento necessários.
var requiredTraining = []trainingData{
	{fileName: "grc.traineddata", url: "https://github.com/tesseract-ocr/tessdata_best/raw/main/grc.traineddata"},
	{fileName: "lat.traineddata", url: "https://github.com/tesseract-ocr/tessdata_best/raw/main/lat.traineddata"},
	{fileName: "heb.traineddata", url: "https://github.com/tesseract-ocr/tessdata_best/raw/main/heb.traineddata"},
}

// ensureTrainingData verifica se os arquivos de treinamento existem e, se não, faz o download.
func ensureTrainingData() error {
	var filesToDownload []trainingData

	// Verifica quais arquivos estão faltando
	for _, data := range requiredTraining {
		filePath := filepath.Join(trainingDir, data.fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			filesToDownload = append(filesToDownload, data)
		}
	}

	// Se não houver arquivos faltando, retorna
	if len(filesToDownload) == 0 {
		fmt.Println("Arquivos de treinamento já existem. Nenhuma ação necessária.")
		return nil
	}

	fmt.Println("Alguns arquivos de treinamento estão faltando. Iniciando download...")

	// Cria o diretório de treinamento se não existir
	if err := os.MkdirAll(trainingDir, os.ModePerm); err != nil {
		return fmt.Errorf("falha ao criar o diretório de treinamento: %w", err)
	}

	// Baixa os arquivos que estão faltando
	for _, data := range filesToDownload {
		fmt.Printf("Baixando %s...\n", data.fileName)
		filePath := filepath.Join(trainingDir, data.fileName)
		if err := downloadFile(filePath, data.url); err != nil {
			log.Printf("Erro ao baixar %s: %v", data.fileName, err)
		}
	}
	fmt.Println("Downloads concluídos!")
	return nil
}

// downloadFile é uma função auxiliar para baixar um arquivo de uma URL.
func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status de resposta ruim: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// processPDF envia um arquivo PDF para o Tika e salva o texto extraído.
func processPDF(pdfPath string, wg *sync.WaitGroup, client *http.Client) {
	defer wg.Done()

	log.Printf("Processando arquivo: %s\n", pdfPath)

	file, err := os.Open(pdfPath)
	if err != nil {
		log.Printf("Erro ao abrir %s: %v\n", pdfPath, err)
		return
	}
	defer file.Close()

	req, err := http.NewRequest("PUT", tikaServerURL, file)
	if err != nil {
		log.Printf("Erro ao criar requisição para %s: %v\n", pdfPath, err)
		return
	}
	req.Header.Set("Accept", "text/plain; charset=utf-8")
	req.Header.Set("Content-Type", "application/pdf")
	req.Header.Set("X-Tika-PDFOcrStrategy", "ocr_and_text")
	req.Header.Set("X-Tika-OCRLanguage", "grc+lat+heb+por+eng")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro ao enviar requisição para Tika (%s): %v\n", pdfPath, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Tika retornou um status não-OK para %s: %s\n", pdfPath, resp.Status)
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Corpo da resposta do Tika: %s\n", string(bodyBytes))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Erro ao ler resposta do Tika para %s: %v\n", pdfPath, err)
		return
	}

	baseName := filepath.Base(pdfPath)
	fileNameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// Salva o arquivo .txt
	txtPath := filepath.Join(resultsDir, fileNameWithoutExt+".txt")
	err = os.WriteFile(txtPath, body, 0644)
	if err != nil {
		log.Printf("Erro ao salvar arquivo de resultado %s: %v\n", txtPath, err)
	} else {
		log.Printf("Texto extraído de %s e salvo em %s\n", pdfPath, txtPath)
	}

	// Salva o arquivo .docx usando a nova biblioteca gratuita
	docxPath := filepath.Join(resultsDir, fileNameWithoutExt+".docx")
	doc := document.New()
	para := doc.AddParagraph()
	para.AddRun().AddText(string(body))

	err = doc.SaveToFile(docxPath)
	if err != nil {
		log.Printf("Erro ao salvar arquivo Word %s: %v\n", docxPath, err)
	} else {
		log.Printf("Arquivo Word salvo em %s\n", docxPath)
	}
}

func main() {
	if err := ensureTrainingData(); err != nil {
		log.Fatalf("Falha na verificação dos dados de treinamento: %v", err)
	}

	if err := os.MkdirAll(pdfsDir, os.ModePerm); err != nil {
		log.Fatalf("Falha ao criar o diretório '%s': %v", pdfsDir, err)
	}
	if err := os.MkdirAll(resultsDir, os.ModePerm); err != nil {
		log.Fatalf("Falha ao criar o diretório '%s': %v", resultsDir, err)
	}

	files, err := os.ReadDir(pdfsDir)
	if err != nil {
		log.Fatalf("Falha ao ler o diretório de PDFs: %v", err)
	}

	var pdfFiles []fs.DirEntry
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".pdf") {
			pdfFiles = append(pdfFiles, file)
		}
	}

	if len(pdfFiles) == 0 {
		log.Println("A pasta 'pdfs' está vazia ou não contém arquivos .pdf.")
		log.Println("Por favor, adicione arquivos PDF para processamento.")
		return
	}

	var wg sync.WaitGroup
	client := &http.Client{}

	for _, file := range pdfFiles {
		pdfPath := filepath.Join(pdfsDir, file.Name())
		wg.Add(1)
		go processPDF(pdfPath, &wg, client)
	}

	wg.Wait()
	log.Println("Processamento de todos os arquivos concluído.")
}
