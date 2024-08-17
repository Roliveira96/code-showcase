package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func main() {
	// Cria um novo arquivo Excel
	f := excelize.NewFile()

	// Usa nomes de planilhas
	mainSheet := "A ser preenchido"

	if _, err := f.NewSheet(mainSheet); err != nil {
		fmt.Println(err)
		return
	}

	// Define o nome da planilha principal como a primeira planilha

	index, err := f.GetSheetIndex(mainSheet)
	if err != nil {
		fmt.Println(err)
		return
	}
	f.SetActiveSheet(index)

	// Cria cabeçalhos para as colunas na planilha principal
	f.SetCellValue(mainSheet, "A1", "Número")
	f.SetCellValue(mainSheet, "B1", "Mês")
	f.SetCellValue(mainSheet, "C1", "Ano")

	// Adiciona os meses na planilha de configuração para a validação de dados
	months := []string{"Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho", "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"}
	/*	for i, month := range months {
		f.SetCellValue(configSheet, fmt.Sprintf("A%d", i+1), month)
	}*/

	// Define a validação de dados para a coluna "Mês" (B2:B1000) usando uma lista de strings
	dvMonth := excelize.NewDataValidation(true)
	dvMonth.Sqref = "B2:B1000"
	dvMonth.SetDropList(months)
	if err := f.AddDataValidation(mainSheet, dvMonth); err != nil {
		fmt.Println(err)
		return
	}

	// Define a validação de dados para a coluna "Ano" (C2:C1000) com anos de 2010 a 2030
	years := []string{}
	for year := 2010; year <= 2030; year++ {
		years = append(years, fmt.Sprintf("%d", year))
	}

	// Define a validação de dados para a coluna "Ano" (C2:C1000) com anos de 2010 a 2030
	dvAno := excelize.NewDataValidation(true)
	dvAno.Sqref = "C2:C1000"
	dvAno.SetDropList(years)
	if err := f.AddDataValidation(mainSheet, dvAno); err != nil {
		fmt.Println(err)
		return
	}

	if err := f.DeleteSheet("Sheet1"); err != nil {
		fmt.Println("Erro ao remover a planilha:", err)
		return
	}

	// Salva o arquivo Excel
	if err := f.SaveAs("example.xlsx"); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Arquivo Excel 'example.xlsx' criado com sucesso.")
}
