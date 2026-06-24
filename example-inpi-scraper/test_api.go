package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type BrasilAPIResponse struct {
	Cnpj        string `json:"cnpj"`
	RazaoSocial string `json:"razao_social"`
	Email       string `json:"email"`
	DddTelef1   string `json:"ddd_telefone_1"`
	Telefone1   string `json:"telefone_1"`
}

func main() {
	cnpj := "35520776000117" // INOVA ARQUITETURA
	reqUrl := "https://casadosdados.com.br/solucao/cnpj/" + cnpj
	
	req, _ := http.NewRequest("GET", reqUrl, nil)
	req.Header.Set("User-Agent", "Bot-Pesquisa-Empresarial-INPI")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)

	var apiResp BrasilAPIResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)
	
	fmt.Printf("Empresa: %s\nEmail: %s\nTelefone: (%s) %s\n", apiResp.RazaoSocial, apiResp.Email, apiResp.DddTelef1, apiResp.Telefone1)
}
