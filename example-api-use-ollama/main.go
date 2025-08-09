package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath" // Importe para manipulação de caminhos

	"github.com/joho/godotenv" // Importe o godotenv
)

type RequestBody struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ResponseBody struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}

type Config struct {
	APIPort   string
	OllamaURI string
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Erro ao obter o diretório atual: %v", err)
	}

	envPath := filepath.Join(dir, ".env")

	err = godotenv.Load(envPath)
	if err != nil {
		log.Printf("Aviso: Não foi possível carregar o arquivo .env em '%s'. Usando variáveis de ambiente do sistema ou padrões.", envPath)
	}

	config := loadConfig()
	mux := http.NewServeMux()

	mux.HandleFunc("/ask-ollama", func(w http.ResponseWriter, r *http.Request) {
		handleAskOllama(w, r, config.OllamaURI)
	})

	log.Printf("Servidor iniciado na porta %s...\n", config.APIPort)
	log.Fatal(http.ListenAndServe(":"+config.APIPort, mux))
}

func loadConfig() Config {
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		log.Println("Variável de ambiente API_PORT não encontrada. Usando a porta padrão 8080.")
		apiPort = "8080"
	}

	ollamaURI := os.Getenv("OLLAMA_URI")
	if ollamaURI == "" {
		log.Println("Variável de ambiente OLLAMA_URI não encontrada. Usando URI padrão http://localhost:11434/api/generate.")
		ollamaURI = "http://localhost:11434/api/generate"
	}

	return Config{
		APIPort:   apiPort,
		OllamaURI: ollamaURI,
	}
}

func handleAskOllama(w http.ResponseWriter, r *http.Request, ollamaURI string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}

	var requestBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "JSON da requisição inválido.", http.StatusBadRequest)
		return
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		http.Error(w, "Erro ao converter para JSON.", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(ollamaURI, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("Erro ao fazer a requisição para o Ollama: %v", err)
		http.Error(w, "Erro ao se comunicar com o Ollama.", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Erro ao ler a resposta do Ollama: %v", err)
		http.Error(w, "Erro ao ler a resposta do Ollama.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}
