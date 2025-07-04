package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// TokenResponse representa a estrutura da resposta do token do Azure AD.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	ExtExpiresIn int    `json:"ext_expires_in"`
}

// EmailAddress representa um endereço de e-mail.
type EmailAddress struct {
	Address string `json:"address"`
}

// Recipient representa um destinatário de e-mail.
type Recipient struct {
	EmailAddress EmailAddress `json:"emailAddress"`
}

// EmailBodyContent representa o conteúdo do corpo do e-mail.
type EmailBodyContent struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

// InternetMessageHeader representa um cabeçalho de mensagem personalizado.
type InternetMessageHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Message representa a estrutura principal da mensagem de e-mail.
type Message struct {
	Subject                string                  `json:"subject"`
	Body                   EmailBodyContent        `json:"body"`
	ToRecipients           []Recipient             `json:"toRecipients"`
	InternetMessageHeaders []InternetMessageHeader `json:"internetMessageHeaders,omitempty"`
}

// EmailMessage representa a estrutura completa para envio de e-mail.
type EmailMessage struct {
	Message         Message `json:"message"`
	SaveToSentItems bool    `json:"saveToSentItems"`
}

// tokenCache armazena o token de acesso e seu tempo de expiração.
var (
	tokenCache struct {
		AccessToken string
		Expiry      time.Time
	}
	mu sync.Mutex // Mutex para proteger o acesso a tokenCache
)

// CreateEmailBody cria uma estrutura EmailBodyContent.
func CreateEmailBody(contentType, content string) EmailBodyContent {
	return EmailBodyContent{
		ContentType: contentType,
		Content:     content,
	}
}

// CreateToRecipients cria uma fatia de Recipient a partir de uma lista de e-mails.
func CreateToRecipients(emails ...string) []Recipient {
	recipients := make([]Recipient, len(emails))
	for i, email := range emails {
		recipients[i] = Recipient{
			EmailAddress: EmailAddress{Address: email},
		}
	}
	return recipients
}

// CreateEmailMessage cria uma estrutura EmailMessage completa com cabeçalhos personalizados.
func CreateEmailMessage(subject string, body EmailBodyContent, toRecipients []Recipient, saveToSentItems bool, customHeaders map[string]string) EmailMessage {
	msg := Message{
		Subject:      subject,
		Body:         body,
		ToRecipients: toRecipients,
	}

	if len(customHeaders) > 0 {
		msg.InternetMessageHeaders = make([]InternetMessageHeader, 0, len(customHeaders))
		for name, value := range customHeaders {
			msg.InternetMessageHeaders = append(msg.InternetMessageHeaders, InternetMessageHeader{Name: name, Value: value})
		}
	}

	return EmailMessage{
		Message:         msg,
		SaveToSentItems: saveToSentItems,
	}
}

// loadEnv carrega as variáveis de ambiente de um arquivo .env.
func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Erro ao carregar arquivo .env:", err)
	}
}

// getAccessToken obtém um novo token de acesso do Azure AD.
func getAccessToken(tenantID, clientID, clientSecret string) (string, int, error) {
	authURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)

	data := fmt.Sprintf("client_id=%s&scope=https://graph.microsoft.com/.default&client_secret=%s&grant_type=client_credentials",
		clientID, clientSecret)

	req, err := http.NewRequest("POST", authURL, bytes.NewBufferString(data))
	if err != nil {
		return "", 0, fmt.Errorf("erro ao criar requisição de token: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("erro ao enviar requisição de token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("erro na resposta do token - status: %d, corpo: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("erro ao ler corpo da resposta do token: %w", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return "", 0, fmt.Errorf("erro ao decodificar resposta do token: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", 0, fmt.Errorf("token de acesso vazio na resposta")
	}

	fmt.Println("Token de acesso obtido com sucesso.")
	return tokenResp.AccessToken, tokenResp.ExpiresIn, nil
}

// getValidAccessToken gerencia o cache e a renovação do token de acesso.
func getValidAccessToken(tenantID, clientID, clientSecret string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	if tokenCache.AccessToken != "" && time.Now().Before(tokenCache.Expiry.Add(-5*time.Minute)) {
		fmt.Println("Usando token de acesso em cache.")
		return tokenCache.AccessToken, nil
	}

	fmt.Println("Obtendo ou renovando token de acesso...")
	accessToken, expiresIn, err := getAccessToken(tenantID, clientID, clientSecret)
	if err != nil {
		return "", err
	}

	tokenCache.AccessToken = accessToken
	tokenCache.Expiry = time.Now().Add(time.Duration(expiresIn) * time.Second)

	fmt.Printf("Novo token de acesso obtido. Expira em %d segundos (%s a partir de agora).\n", expiresIn, time.Duration(expiresIn)*time.Second)
	return tokenCache.AccessToken, nil
}

// sendOutlookEmail envia um e-mail usando a Microsoft Graph API.
func sendOutlookEmail(accessToken, sender string, emailData EmailMessage) error {
	graphAPIURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/sendMail", sender)

	jsonBody, err := json.Marshal(emailData)
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

	if resp.StatusCode != http.StatusAccepted {
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
	idRicardo := os.Getenv("IDAUTH")

	if tenantID == "" || clientID == "" || clientSecret == "" || senderEmail == "" || recipientEmail == "" {
		fmt.Println("Erro: Verifique se todas as variáveis no arquivo .env estão preenchidas.")
		os.Exit(1)
	}

	// Usa getValidAccessToken para obter ou renovar o token
	accessToken, err := getValidAccessToken(tenantID, clientID, clientSecret)
	if err != nil {
		fmt.Println("Erro ao obter/validar token de acesso:", err)
		os.Exit(1)
	}

	emailSubject := "Teste de E-mail da Aplicação - Grownt.tech (Golang - Refatorado)"
	emailBodyContentHTML := `
    <html>
        <body>
            <p>Olá Ricardo,</p>
            <p>Este é um e-mail de teste enviado automaticamente pela sua aplicação em <strong>Golang</strong> utilizando a <strong>Microsoft Graph API</strong> e o <strong>Azure AD</strong>.</p>
            <p>Se você recebeu esta mensagem, a configuração está funcionando!</p>
            <br>
            <p>Este e-mail demonstra o código <strong>refatorado</strong> com funções para cada responsabilidade.</p>
            <br>
            <p>Atenciosamente,</p>
            <p>Equipe de Suporte Grownt.tech</p>
        </body>
    </html>
    `
	emailBody := CreateEmailBody("HTML", emailBodyContentHTML)

	toRecipients := CreateToRecipients(recipientEmail)

	customHeaders := map[string]string{
		"X-Sent-Via":     "Mensageria",
		"X-App-Name":     "Boards",
		"X-ID-SEND-AUTH": idRicardo,
	}

	emailMessage := CreateEmailMessage(emailSubject, emailBody, toRecipients, true, customHeaders)

	fmt.Printf("Tentando enviar e-mail de %s para %s...\n", senderEmail, recipientEmail)
	err = sendOutlookEmail(accessToken, senderEmail, emailMessage)
	if err != nil {
		fmt.Println("Ocorreu um erro no envio do e-mail:", err)
		os.Exit(1)
	}

	fmt.Println("Processo de envio finalizado.")
}
