package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"golang.org/x/oauth2"
)

// Configuração do Keycloak
var (
	clientID    = "frontend"
	redirectURL = "http://localhost:5040/callback"
	keycloakURL = "http://localhost:9080/realms/plataformagt"
)

var (
	authConfig = &oauth2.Config{
		ClientID: clientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:  keycloakURL + "/protocol/openid-connect/auth",
			TokenURL: keycloakURL + "/protocol/openid-connect/token",
		},
		RedirectURL: redirectURL,
		Scopes:      []string{"openid", "profile", "email"},
	}
	globalToken *oauth2.Token
)

func main() {
	// Inicia um servidor HTTP para escutar o callback
	http.HandleFunc("/callback", handleCallback)
	go http.ListenAndServe(":5040", nil)

	// Abre o navegador para iniciar o fluxo de autenticação
	authURL := authConfig.AuthCodeURL("state")
	openbrowser(authURL)

	// Aguarda o token ser obtido
	for globalToken == nil {
	}

	// Usa o token para fazer requisições
	fmt.Println("Token:", globalToken.AccessToken)

	req, err := http.NewRequest("GET", "http://localhost:8082/api/squads/39", nil)
	if err != nil {
		log.Fatal(err)
	}

	// Adiciona o token de acesso ao cabeçalho da requisição
	req.Header.Add("Authorization", "Bearer "+globalToken.AccessToken)

	// Envia a requisição
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Lê o corpo da resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Imprime o JSON de retorno
	fmt.Println(string(body))
	// ...
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Código de autorização não encontrado", http.StatusBadRequest)
		return
	}

	// Troca o código pelo token
	token, err := authConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Erro ao trocar o código pelo token", http.StatusInternalServerError)
		return
	}

	globalToken = token

	// Exibe uma mensagem de sucesso e fecha a janela do navegador
	fmt.Fprintln(w, "Autenticação concluída com sucesso! Você pode fechar esta janela.")
}

// Função para abrir o navegador padrão
func openbrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("sistema operacional não suportado")
	}
	if err != nil {
		log.Fatal(err)
	}
}
