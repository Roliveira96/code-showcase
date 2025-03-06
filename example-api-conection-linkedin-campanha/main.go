package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	// Obter o token de acesso do ambiente (substitua pelo nome da sua variável de ambiente)
	var accessToken = "AQWH6Uf_9zCHw0XZnGeKUOublmXeCA216gY9DVdLhFg6J5nPqRGDEBGHfKo9y8se2SAJDWjyfyY2ATn3AKteush0RSJ4CGvV-60W8bUnwt0lfh-qG059f6cpr9ih1ULmzueIDIxMykTDnhErnHTOMpPryJSiGoVvPm_wBlAtDV7sY2aR-dIv8GkiBPXwL7Ez27Is2W-GteGbZ0PuHWGtiAwnOe4a2GJlTZk-GdD7KuoF73-Z83rziy0J6mDZQxsBp6aKGl1wZ72eTHFZu9Uu_zgscEIJ2AZcpN-Tjg7KmRuP1j8DQ8GHb77HPGHwmZnyMCfEXMZpUdUc1Gnh1xoDTo4XEBNxiA"

	// IDs das campanhas a serem buscadas
	campaignIDs := []string{
		"urn:li:sponsoredCampaign:318821733", // Substitua pelos IDs reais das suas campanhas
	}

	// Construir o corpo da requisição BATCH_GET
	var requests []map[string]interface{}
	for _, id := range campaignIDs {
		requests = append(requests, map[string]interface{}{
			"id":     id,
			"method": "GET",
			"url":    fmt.Sprintf("/v2/adCampaignsV2/%s", id),
		})
	}

	requestBody, err := json.Marshal(map[string]interface{}{"requests": requests})
	if err != nil {
		log.Fatal("Erro ao construir o corpo da requisição:", err)
	}

	// Criar a requisição
	client := &http.Client{}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "https://api.linkedin.com/v2/adCampaignsV2?action=BATCH_GET", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatal("Erro ao criar a requisição:", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")

	// Fazer a requisição
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Erro ao buscar as campanhas:", err)
	}
	defer resp.Body.Close()

	// Ler a resposta
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Erro ao ler a resposta:", err)
	}

	// Analisar a resposta JSON
	var batchResponse []map[string]interface{}
	err = json.Unmarshal(body, &batchResponse)
	if err != nil {
		log.Fatal("Erro ao analisar a resposta JSON:", err)
	}

	// Imprimir os nomes das campanhas
	fmt.Println("Campanhas:")
	for _, response := range batchResponse {
		if response["status"] != nil {
			status := response["status"].(float64)
			if status == 200 { // Verificar se a requisição individual foi bem-sucedida
				if response["value"] != nil {
					value := response["value"].(map[string]interface{})
					if value["campaignGroup"] != nil {
						campaignGroup := value["campaignGroup"].(map[string]interface{})
						if campaignGroup["name"] != nil {
							fmt.Println("- ", campaignGroup["name"])
						}
					}
				}
			} else {
				fmt.Println("Erro ao buscar campanha:", response["headers"])
			}
		}
	}
}
