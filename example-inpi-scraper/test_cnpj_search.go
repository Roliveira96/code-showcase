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
	reqUrl := "https://cnpj.biz/procurar?q=" + url.QueryEscape(razao)

	req, _ := http.NewRequest("GET", reqUrl, nil)
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

	fmt.Println("Body contains match:", len(html))
	// Look for 14-digit pattern
	fmt.Println("Sample HTML:", html[0:500])
}
