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
	busca := "https://www.ecosia.org/search?q=" + url.QueryEscape(termo)

	req, _ := http.NewRequest("GET", busca, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)
	bodyBytes, _ := io.ReadAll(resp.Body)
	html := string(bodyBytes)

	fmt.Println("Body length:", len(html))
	// Look for CNPJ patterns
	fmt.Println("Sample HTML:", html[100:600])
}
