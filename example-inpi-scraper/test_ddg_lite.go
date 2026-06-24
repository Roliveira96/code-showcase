package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func main() {
	razao := "NOVACARE CURSOS E TREINAMENTOS LTDA"
	termo := fmt.Sprintf("CNPJ \"%s\"", razao)
	busca := "https://duckduckgo.com/lite/?q=" + url.QueryEscape(termo)

	req, _ := http.NewRequest("GET", busca, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Body length:", len(body))
	fmt.Println("Sample:", string(body)[0:500])
}
