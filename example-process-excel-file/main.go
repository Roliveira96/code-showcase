package main

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"strings"
)

type Employee struct {
	Name      string `json:"nome_colaborador"`
	Position  string `json:"cargo"`
	CPF       string `json:"cpf"`
	BirthDate string `json:"data_nascimento"`
	Phone     string `json:"telefone"`
	Active    bool   `json:"ativo"`
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
	var employees []Employee

	if len(rows) == 0 {
		panic("O arquivo Excel está vazio.")
	}

	colunas := rows[0]

	for i, row := range rows {
		if i == 0 {
			continue
		}

		var employee Employee

		for j, cell := range row {
			nomeColuna := strings.ToLower(strings.ReplaceAll(colunas[j], " ", "_"))
			switch nomeColuna {
			case "nome_colaborador":
				employee.Name = cell
			case "data_de_nascimento":
				employee.BirthDate = cell
			case "telefone":
				employee.Phone = cell
			case "cargo":
				employee.Position = cell
			case "ativo":
				employee.Active = isActive(cell)
			}
		}

		employees = append(employees, employee)
	}

	for _, employee := range employees {

		active := "Não"
		if employee.Active {
			active = "Sim"
		}

		fmt.Printf("Nome: %s, data nascimento: %s, telefone: %s, cargo: %s, ativo: %s\n", employee.Name, employee.BirthDate, employee.Phone, employee.Position, active)
	}
}

func isActive(texto string) bool {

	texto = strings.ToLower(texto)
	if texto == "sim" {
		return true
	}
	return false
}
