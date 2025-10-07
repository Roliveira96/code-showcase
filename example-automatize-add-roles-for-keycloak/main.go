package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Estruturas para a API do Keycloak
type RolePayload struct {
	Name        string `json:"name"`
	Composite   bool   `json:"composite"`
	ClientRole  bool   `json:"clientRole"`
	Description string `json:"description"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func main() {
	// Carrega o .env
	err := godotenv.Load()
	if err != nil {
		fmt.Println("ERRO: Não foi possível carregar o arquivo .env. Certifique-se de que ele existe.")
		os.Exit(1)
	}

	// 1. Obter o Access Token
	token, err := getAccessToken()
	if err != nil {
		fmt.Printf("ERRO ao obter o token de acesso: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Token de acesso obtido com sucesso.")

	// 2. Ler as roles do arquivo
	roles, err := readRolesFromFile(os.Getenv("ROLES_FILE"))
	if err != nil {
		fmt.Printf("ERRO ao ler o arquivo de roles: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total de %d roles encontradas.\n", len(roles))

	// 3. Criar as roles no Keycloak
	for _, roleName := range roles {
		createRole(token, roleName)
	}
}

// getAccessToken obtém um token de acesso usando credenciais de Admin
func getAccessToken() (string, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		os.Getenv("KC_URL"),
		os.Getenv("KC_ADMIN_REALM"))

	data := strings.NewReader(fmt.Sprintf(
		"client_id=%s&username=%s&password=%s&grant_type=password",
		os.Getenv("KC_CLIENT_ID"),
		os.Getenv("KC_USERNAME"),
		os.Getenv("KC_PASSWORD"),
	))

	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("falha ao autenticar, status: %d, resposta: %s", resp.StatusCode, body)
	}

	var tokenRes TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return "", err
	}

	return tokenRes.AccessToken, nil
}

// readRolesFromFile lê as roles de um arquivo de texto, uma por linha
func readRolesFromFile(filename string) ([]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var roles []string
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			roles = append(roles, trimmedLine)
		}
	}
	return roles, nil
}

// createRole tenta criar a role no Keycloak
func createRole(token, roleName string) {
	url := fmt.Sprintf("%s/admin/realms/%s/roles",
		os.Getenv("KC_URL"),
		os.Getenv("KC_TARGET_REALM"))

	payload := RolePayload{
		Name:        roleName,
		Composite:   false,
		ClientRole:  false,
		Description: fmt.Sprintf("Role para o time: %s", roleName),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("ERRO: Falha ao serializar payload para %s: %v\n", roleName, err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Printf("ERRO: Falha ao criar requisição para %s: %v\n", roleName, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("ERRO: Falha na requisição HTTP para %s: %v\n", roleName, err)
		return
	}
	defer resp.Body.Close()

	// O Keycloak retorna 201 (Created) para sucesso
	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("✅ Role '%s' criada com sucesso.\n", roleName)
	} else if resp.StatusCode == http.StatusConflict || resp.StatusCode == http.StatusNoContent {
		// Keycloak retorna 409 (Conflict) se a role já existe
		fmt.Printf("⚠️ Role '%s' já existe. Pulando a criação.\n", roleName)
	} else {
		// Tratar outros erros
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ ERRO ao criar a role '%s'. Status: %d. Resposta: %s\n", roleName, resp.StatusCode, body)
	}
}
