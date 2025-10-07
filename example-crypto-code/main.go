package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// Esta é a chave de criptografia HARDCODED.
// PARA AMBIENTES DE PRODUÇÃO, ISSO É ALTAMENTE INSEGURO E NUNCA DEVE SER FEITO.
// Em um ambiente real, a chave deve vir de um KMS (Key Management Service)
// ou de uma variável de ambiente segura.
// comando para gerar encryptionKeyBase64 $ openssl rand -base64 32
const encryptionKeyBase64 = "5BEJo8E4JAyUNdGQD8EUj3Q2gMx7N0r5NcpvQ+V8AHQ=" // Exemplo de chave AES-256 (32 bytes) em Base64

// getEncryptionKey agora simplesmente decodifica a chave hardcoded.
func getEncryptionKey() ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(encryptionKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("falha ao decodificar chave de criptografia hardcoded: %w", err)
	}
	if len(key) != 32 { // AES-256 requer uma chave de 32 bytes
		return nil, errors.New("chave de criptografia deve ter 32 bytes (256 bits) após decodificação base64")
	}
	return key, nil
}

// encrypt data using AES-256 GCM.
// Retorna os dados criptografados como uma string Base64.
func encrypt(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("falha ao criar cipher AES: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("falha ao criar GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("falha ao gerar nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt data using AES-256 GCM.
// Recebe os dados criptografados em Base64 e retorna o plaintext.
func decrypt(encryptedDataBase64 string, key []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedDataBase64)
	if err != nil {
		return nil, fmt.Errorf("falha ao decodificar dados criptografados: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar cipher AES: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext muito curto para conter nonce")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("falha ao descriptografar (dados adulterados ou chave incorreta?): %w", err)
	}
	return plaintext, nil
}

func main() {
	encryptionKey, err := getEncryptionKey()
	if err != nil {
		fmt.Println("Erro ao obter chave de criptografia:", err)
		fmt.Println("Por favor, verifique se a chave hardcoded é válida e tem o tamanho correto.")
		return
	}

	phraseOriginal := "Esta é uma frase muito secreta que precisa ser protegida!"
	fmt.Println("Frase original:", phraseOriginal)

	encryptedPhrase, err := encrypt([]byte(phraseOriginal), encryptionKey)
	if err != nil {
		fmt.Println("Erro ao criptografar a frase:", err)
		return
	}
	fmt.Println("Frase criptografada (Base64):", encryptedPhrase)

	decryptedPhraseBytes, err := decrypt(encryptedPhrase, encryptionKey)
	if err != nil {
		fmt.Println("Erro ao descriptografar a frase:", err)
		return
	}
	decryptedPhrase := string(decryptedPhraseBytes)
	fmt.Println("Frase descriptografada:", decryptedPhrase)

	if phraseOriginal == decryptedPhrase {
		fmt.Println("SUCESSO: A frase original e a descriptografada são idênticas!")
	} else {
		fmt.Println("ERRO: A frase original e a descriptografada NÃO coincidem!")
	}

	fmt.Println("\n--- Importante ---")
	fmt.Println("AVISO DE SEGURANÇA: Este exemplo usa uma chave de criptografia hardcoded APENAS para demonstração em playgrounds.")
	fmt.Println("NUNCA utilize chaves hardcoded ou armazenadas diretamente no código em ambientes de produção ou para dados sensíveis.")
	fmt.Println("Sempre use um Key Management Service (KMS) ou variáveis de ambiente seguras para as chaves em produção.")
}
