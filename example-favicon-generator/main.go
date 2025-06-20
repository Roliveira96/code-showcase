package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FaviconSpec defines the filename and dimensions for a single favicon.
type FaviconSpec struct {
	Filename string
	Width    int
	Height   int
}

// resizeImage uses an external tool (like ImageMagick's 'convert') to resize.
// IMPORTANT: Requires ImageMagick or GraphicsMagick to be installed and in the PATH.
func resizeImage(inputPath, outputPath string, width, height int) error {
	// Ensure the output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		// Don't return here, maybe it exists, let convert try
		log.Printf("Aviso: não foi possível garantir a criação do diretório %s: %v", outputDir, err)
	}

	cmd := "convert" // Assumes ImageMagick's convert command
	args := []string{
		inputPath,
		"-resize",
		fmt.Sprintf("%dx%d!", width, height), // Add '!' to force exact dimensions
		outputPath,
	}

	command := exec.Command(cmd, args...)
	// Capture standard error for better debugging
	output, err := command.CombinedOutput()
	if err != nil {
		// Log the command output for debugging if there's an error
		log.Printf("Erro ao executar o comando: %s %s", cmd, strings.Join(args, " "))
		log.Printf("Saída do comando:\n%s", string(output))
		return fmt.Errorf("erro ao redimensionar '%s' para %s usando '%s': %w", filepath.Base(inputPath), filepath.Base(outputPath), cmd, err)
	}
	return nil
}

// generateFavicons orchestrates the creation of all specified favicons.
func generateFavicons(inputPath string, specs []FaviconSpec) error {
	outputDir := "favicon"
	err := os.MkdirAll(outputDir, 0755) // 0755 gives read/write/execute to owner, read/execute to others
	if err != nil {
		return fmt.Errorf("erro ao criar diretório '%s': %w", outputDir, err)
	}

	fmt.Printf("Iniciando geração de favicons a partir de: %s\n", inputPath)

	var errorsOccurred bool = false
	for _, spec := range specs {
		outputPath := filepath.Join(outputDir, spec.Filename)
		fmt.Printf("Gerando %s (%dx%d)...\n", outputPath, spec.Width, spec.Height)

		if err := resizeImage(inputPath, outputPath, spec.Width, spec.Height); err != nil {
			// Log the specific error and continue with the next favicon
			log.Printf("ERRO ao gerar %s: %v", outputPath, err)
			errorsOccurred = true // Mark that at least one error happened
			continue              // Skip to the next iteration
		}
	}

	if errorsOccurred {
		fmt.Println("\nAtenção: Alguns favicons não puderam ser gerados devido a erros (ver logs acima).")
		// Return an error to indicate partial failure, although the process completed.
		return fmt.Errorf("processo de geração de favicons concluído com erros")
	}

	fmt.Println("\nFavicons gerados com sucesso na pasta 'favicon'.")
	return nil
}

func main() {
	// 1. Define the target favicon specifications
	faviconSpecs := []FaviconSpec{
		{Filename: "apple-icon-57x57.png", Width: 57, Height: 57},
		{Filename: "apple-icon-60x60.png", Width: 60, Height: 60},
		{Filename: "apple-icon-72x72.png", Width: 72, Height: 72},
		{Filename: "apple-icon-76x76.png", Width: 76, Height: 76},
		{Filename: "apple-icon-114x114.png", Width: 114, Height: 114},
		{Filename: "apple-icon-120x120.png", Width: 120, Height: 120},
		{Filename: "apple-icon-144x144.png", Width: 144, Height: 144},
		{Filename: "apple-icon-152x152.png", Width: 152, Height: 152},
		{Filename: "apple-icon-180x180.png", Width: 180, Height: 180},
		{Filename: "android-icon-192x192.png", Width: 192, Height: 192},
		{Filename: "favicon-32x32.png", Width: 32, Height: 32},
		{Filename: "favicon-96x96.png", Width: 96, Height: 96},
		{Filename: "favicon-16x16.png", Width: 16, Height: 16},
		// Note: ms-icon-144x144.png is the same size as apple-icon-144x144.png
		// If you need a separate file, just add it. If the content is identical,
		// resizing twice is slightly inefficient but harmless. If you want to be
		// efficient, you could copy the file after generating apple-icon-144x144.png.
		// For simplicity, we regenerate it here.
		{Filename: "ms-icon-144x144.png", Width: 144, Height: 144},
	}

	// 2. Get input image path from command line arguments
	if len(os.Args) != 2 {
		fmt.Println("Uso: go run main.go <caminho_para_sua_imagem.png_ou_jpg>")
		fmt.Println("Exemplo: go run main.go logo.png")
		return // Exit if incorrect number of arguments
	}
	inputPath := os.Args[1]

	// 3. Validate input file
	// Check if file exists
	fileInfo, err := os.Stat(inputPath)
	if os.IsNotExist(err) {
		log.Fatalf("Erro: Arquivo de entrada não encontrado em '%s'", inputPath)
	}
	if err != nil { // Handle other potential stat errors
		log.Fatalf("Erro ao verificar o arquivo de entrada '%s': %v", inputPath, err)
	}
	if fileInfo.IsDir() {
		log.Fatalf("Erro: O caminho de entrada '%s' é um diretório, não um arquivo.", inputPath)
	}

	// Check file format (extension)
	ext := strings.ToLower(filepath.Ext(inputPath))
	if ext != ".png" && ext != ".jpeg" && ext != ".jpg" {
		log.Fatalf("Erro: Formato de arquivo não suportado: '%s'. Use apenas arquivos PNG ou JPEG.", ext)
	}

	// 4. Check for external dependency ('convert' command)
	_, err = exec.LookPath("convert")
	if err != nil {
		log.Fatalf("Erro: Comando 'convert' (ImageMagick) não encontrado no PATH do sistema. Por favor, instale ImageMagick.")
	}

	// 5. Generate the favicons
	err = generateFavicons(inputPath, faviconSpecs)
	if err != nil {
		// Specific errors during generation are logged inside generateFavicons
		// This final message indicates the overall process finished, possibly with issues.
		log.Printf("Processo finalizado. %v", err) // Use log to potentially capture timestamp/etc.
		os.Exit(1)                                 // Exit with a non-zero code to indicate errors occurred
	}
}

// Note: The decodeImage and encodeImage functions from your example
// are not used in this version because
