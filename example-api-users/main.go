package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Estrutura do usuário
type Usuario struct {
	ID          int    `json:"id"`
	Nome        string `json:"nome"`
	Email       string `json:"email"`
	Localizacao string `json:"localizacao"`
}

func main() {
	// Cria um usuário de exemplo
	usuario := Usuario{
		ID:          1,
		Nome:        "Fulano de Tal",
		Email:       "fulano@exemplo.com",
		Localizacao: "Guarapuava, Paraná",
	}

	// Manipulador do endpoint "/usuario"
	http.HandleFunc("/usuario", func(w http.ResponseWriter, r *http.Request) {
		// Define o cabeçalho Content-Type como "application/json"
		w.Header().Set("Content-Type", "application/json")

		// Codifica o usuário em JSON
		json.NewEncoder(w).Encode(usuario)
	})

	// Verifica se a porta foi fornecida como argumento
	if len(os.Args) != 2 {
		fmt.Println("Uso: ./nomeprograma <porta>")
		return
	}

	// Obtém a porta do primeiro argumento
	porta := ":" + os.Args[1]

	// Inicia o servidor na porta especificada
	fmt.Printf("Servidor iniciado em http://localhost%s/usuario\n", porta)
	log.Fatal(http.ListenAndServe(porta, nil))

	// Cria um arquivo com a porta aberta
	err := ioutil.WriteFile("portas.txt", []byte(porta), 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Porta aberta salva em portas.txt")
}
