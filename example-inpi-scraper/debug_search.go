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
	termo := fmt.Sprintf("cnpj \"%s\"", razao)
	busca := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(termo)

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
	if len(html) < 500 {
		fmt.Println("HTML too short, maybe blocked?")
		fmt.Println(html)
	} else {
		fmt.Println("HTML received. Searching for CNPJ patterns...")
		// Show a bit of the body to see if results are there
		if len(html) > 1000 {
			fmt.Println(html[0:1000])
		}
	}
}
