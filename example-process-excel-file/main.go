package main

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"strings"
)

type Colaborador struct {
	Nome           string `json:"nome_colaborador"`
	Cargo          string `json:"cargo"`
	CPF            string `json:"cpf"`
	DataNascimento string `json:"data_nascimento"`
	Telefone       string `json:"telefone"`
	Ativo          bool   `json:"ativo"`
}

func main() {
	f, err := excelize.OpenFile("./colaboradores.xlsx")
	if err != nil {
		panic(err)
	}

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		sheetName = "Sheet1"
	}

	rows := f.GetRows(sheetName)
	var colaboradores []Colaborador

	if len(rows) == 0 {
		panic("O arquivo Excel está vazio.")
	}

	colunas := rows[0]

	for i, row := range rows {
		if i == 0 {
			continue
		}

		var colaborador Colaborador

		for j, cell := range row {
			nomeColuna := strings.ToLower(strings.ReplaceAll(colunas[j], " ", "_"))
			switch nomeColuna {
			case "nome_colaborador":
				colaborador.Nome = cell
			case "data_de_nascimento":
				colaborador.DataNascimento = cell
			case "telefone":
				colaborador.Telefone = cell
			case "cargo":
				colaborador.Cargo = cell
			case "ativo":
				colaborador.Ativo = isActive(cell)
			}
		}

		colaboradores = append(colaboradores, colaborador)
	}

	for _, colaborador := range colaboradores {
		fmt.Printf("Nome: %s, data nascimento: %s, telefone: %s, cargo: %s, ativo: %t\n", colaborador.Nome, colaborador.DataNascimento, colaborador.Telefone, colaborador.Cargo, colaborador.Ativo)
	}
}

func isActive(texto string) bool {

	texto = strings.ToLower(texto)
	if texto == "sim" {
		return true
	}
	return false
}
