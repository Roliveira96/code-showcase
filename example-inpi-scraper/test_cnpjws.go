package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func main() {
	cnpj := "35520776000117"
	reqUrl := "https://publica.cnpj.ws/cnpj/" + cnpj
	
	req, _ := http.NewRequest("GET", reqUrl, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	
	estabelecimento := data["estabelecimento"].(map[string]interface{})
	fmt.Printf("Email: %v\n", estabelecimento["email"])
	fmt.Printf("Tel1: %v%v\n", estabelecimento["ddd1"], estabelecimento["telefone1"])
}
