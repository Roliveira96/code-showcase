package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"os"
	"time"

	"github.com/joho/godotenv"
)

// TokenResponse representa a estrutura da resposta ao obter um token de acesso
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// EmailMessage representa a estrutura do corpo do e-mail para a Graph API
type EmailMessage struct {
	Message struct {
		Subject string `json:"subject"`
		Body    struct {
			ContentType string `json:"contentType"`
			Content     string `json:"content"`
		} `json:"body"`
		ToRecipients []struct {
			EmailAddress struct {
				Address string `json:"address"`
			} `json:"emailAddress"`
		} `json:"toRecipients"`
	} `json:"message"`
	SaveToSentItems bool `json:"saveToSentItems"`
}

// loadEnv carrega as variáveis de ambiente do arquivo .env
func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Erro ao carregar arquivo .env:", err)
		// Em ambiente de produção, considere fazer o programa falhar ou ter um fallback
	}
}

// getAccessToken obtém um token de acesso usando as credenciais do aplicativo
func getAccessToken(tenantID, clientID, clientSecret string) (string, error) {
	authURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)

	data := fmt.Sprintf("client_id=%s&scope=https://graph.microsoft.com/.default&client_secret=%s&grant_type=client_credentials",
		clientID, clientSecret)

	req, err := http.NewRequest("POST", authURL, bytes.NewBufferString(data))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição de token: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao enviar requisição de token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("erro na resposta do token - status: %d, corpo: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler corpo da resposta do token: %w", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta do token: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("token de acesso vazio na resposta")
	}

	fmt.Println("Token de acesso obtido com sucesso.")
	return tokenResp.AccessToken, nil
}

// sendOutlookEmail envia um e-mail usando a Microsoft Graph API
func sendOutlookEmail(accessToken, sender, recipient, subject, bodyContent string) error {
	graphAPIURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/sendMail", sender)

	emailBody := EmailMessage{
		Message: struct {
			Subject string `json:"subject"`
			Body    struct {
				ContentType string `json:"contentType"`
				Content     string `json:"content"`
			} `json:"body"`
			ToRecipients []struct {
				EmailAddress struct {
					Address string `json:"address"`
				} `json:"emailAddress"`
			} `json:"toRecipients"`
		}{
			Subject: subject,
			Body: struct {
				ContentType string `json:"contentType"`
				Content     string `json:"content"`
			}{
				ContentType: "HTML", // Ou "Text"
				Content:     bodyContent,
			},
			ToRecipients: []struct {
				EmailAddress struct {
					Address string `json:"address"`
				} `json:"emailAddress"`
			}{
				{EmailAddress: struct {
					Address string `json:"address"`
				}{Address: recipient}},
			},
		},
		SaveToSentItems: true,
	}

	jsonBody, err := json.Marshal(emailBody)
	if err != nil {
		return fmt.Errorf("erro ao serializar corpo do e-mail: %w", err)
	}

	req, err := http.NewRequest("POST", graphAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição de envio de e-mail: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao enviar requisição de e-mail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted { // 202 Accepted para sendMail
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("erro ao enviar e-mail - status: %d, corpo: %s", resp.StatusCode, string(bodyBytes))
	}

	fmt.Println("E-mail enviado com sucesso!")
	return nil
}

func main() {
	loadEnv()

	tenantID := os.Getenv("TENANT_ID")
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	senderEmail := os.Getenv("SENDER_EMAIL")
	recipientEmail := os.Getenv("RECIPIENT_EMAIL")

	if tenantID == "" || clientID == "" || clientSecret == "" || senderEmail == "" || recipientEmail == "" {
		fmt.Println("Erro: Verifique se todas as variáveis no arquivo .env estão preenchidas.")
		os.Exit(1)
	}

	// 1. Obter o token de acesso
	accessToken, err := getAccessToken(tenantID, clientID, clientSecret)
	if err != nil {
		fmt.Println("Erro ao obter token de acesso:", err)
		os.Exit(1)
	}

	// 2. Enviar o e-mail
	emailSubject := "Teste de E-mail da Aplicação - Grownt.tech (Golang)"
	emailBodyContent := `
    <html>
        <body>
            <p>Olá Ricardo,</p>
            <p>Este é um e-mail de teste enviado automaticamente pela sua aplicação em <strong>Golang</strong> utilizando a <strong>Microsoft Graph API</strong> e o <strong>Azure AD</strong>.</p>
            <p>Se você recebeu esta mensagem, a configuração está funcionando!</p>
            <br>
            <p>Atenciosamente,</p>
            <p>Equipe de Suporte Grownt.tech</p>
        </body>
    </html>
    `

	fmt.Printf("Tentando enviar e-mail de %s para %s...\n", senderEmail, recipientEmail)
	err = sendOutlookEmail(accessToken, senderEmail, recipientEmail, emailSubject, emailBodyContent)
	if err != nil {
		fmt.Println("Ocorreu um erro no envio do e-mail:", err)
		os.Exit(1)
	}

	fmt.Println("Processo de envio finalizado.")
}
