package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// Usuario define a estrutura para um usuário no payload
type Usuario struct {
	IDUsuario string `json:"idUsuario"`
}

// NotificacaoPayload define a estrutura do corpo da requisição
type NotificacaoPayload struct {
	Tipo       int       `json:"tipo"`
	Prioridade int       `json:"prioridade"`
	Link       string    `json:"link"`
	Titulo     string    `json:"titulo"`
	Conteudo   string    `json:"conteudo"`
	Usuarios   []Usuario `json:"usuarios"`
}

// worker é a função que executa o trabalho de enviar a requisição.
// Vários workers podem rodar ao mesmo tempo.
func worker(id int, wg *sync.WaitGroup, jobs <-chan int, url, token string, jsonData []byte, client *http.Client) {
	// Garante que o WaitGroup será notificado quando este worker terminar
	defer wg.Done()

	// O worker fica "escutando" o canal de jobs.
	// Quando o canal for fechado e esvaziado, o loop termina.
	for reqNum := range jobs {
		fmt.Printf("Worker %d: processando requisição %d...\n", id, reqNum)

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("[Req %d] Erro ao criar requisição: %v", reqNum, err)
			continue // Pega o próximo job
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[Req %d] Erro ao enviar requisição: %v", reqNum, err)
		} else {
			fmt.Printf("[Req %d] Status da Resposta: %s\n", reqNum, resp.Status)
			resp.Body.Close()
		}
	}
}

func main() {
	// Carrega as variáveis do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar o arquivo .env")
	}

	// Pega o token do ambiente
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("A variável de ambiente TOKEN não foi definida no arquivo .env")
	}

	// --- Configurações ---
	url := "http://localhost:8086/api/v1/notificacoes"
	totalRequisicoes := 1000

	// IMPORTANTE: Defina aqui o nível de concorrência.
	// Este é o número de requisições que serão executadas SIMULTANEAMENTE.
	maximoConcorrencia := 50
	// --------------------

	payload := NotificacaoPayload{
		Tipo:       1,
		Prioridade: 2,
		Link:       "https://exemplo.com",
		Titulo:     "Notificação de Exemplo",
		Conteudo:   "Esta é uma notificação de exemplo.",
		Usuarios: []Usuario{
			{IDUsuario: "4a2bf253-f12f-4f01-8f5e-e38f5a2756e1"},
			{IDUsuario: "9dd607a4-0b87-4ba2-b53e-a7bbc90365d5"},
			{IDUsuario: "bf29102c-11df-4a97-a252-5714d4b5a8be"},
			{IDUsuario: "9a1dcdc7-ca98-480e-bba3-6101b5c1379d"},
			{IDUsuario: "a3bbf65b-6b59-4fbf-b3c5-b75ca898a1ee"},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Erro ao converter payload para JSON: %v", err)
	}

	// Criamos um cliente HTTP para ser reutilizado por todos os workers
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Cria um canal para distribuir as "tarefas" (jobs) para os workers
	jobs := make(chan int, totalRequisicoes)

	// Cria um WaitGroup para esperar que todas as goroutines terminem
	var wg sync.WaitGroup

	fmt.Printf("Iniciando %d workers para processar %d requisições...\n", maximoConcorrencia, totalRequisicoes)

	// Inicia os workers
	for w := 1; w <= maximoConcorrencia; w++ {
		wg.Add(1)
		go worker(w, &wg, jobs, url, token, jsonData, client)
	}

	// Adiciona todas as tarefas (de 1 a 3000) no canal de jobs
	for j := 1; j <= totalRequisicoes; j++ {
		jobs <- j
	}
	close(jobs) // Fecha o canal para sinalizar que não há mais tarefas

	// Espera que todos os workers terminem seus trabalhos
	wg.Wait()

	fmt.Println("Todas as requisições foram processadas.")
}
