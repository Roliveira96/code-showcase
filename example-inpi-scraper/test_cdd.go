package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	cnpj := "35520776000117"
	reqUrl := "https://casadosdados.com.br/solucao/cnpj/" + cnpj
	
	req, _ := http.NewRequest("GET", reqUrl, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Final URL:", resp.Request.URL.String())

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Body contains email:", contains(string(body), "M.DOSSO@UOL.COM.BR"))
}

func contains(s, substr string) bool {
	return (len(s) >= len(substr)) && (s[0:len(substr)] == substr || contains(s[1:], substr))
}
