package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"reflect"
	"strconv"
)

var columnLetters = [26]string{
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J",
	"K", "L", "M", "N", "O", "P", "Q", "R", "S", "T",
	"U", "V", "W", "X", "Y", "Z",
}

type Person struct {
	Name  string
	Age   int
	Email string
	Games []string
}

func main() {
	f := excelize.NewFile()
	mainSheet := "A ser preenchido"

	if _, err := f.NewSheet(mainSheet); err != nil {
		fmt.Println(err)
		return
	}

	index, err := f.GetSheetIndex(mainSheet)
	if err != nil {
		fmt.Println(err)
		return
	}
	f.SetActiveSheet(index)

	p := Person{
		Name:  "John Doe",
		Age:   30,
		Email: "john.doe@example.com",
		Games: []string{"Super Mario", "SONIC", "Zelda", "GTA"},
	}

	m := structToMap(p)
	i := uint(0)

	for s, strings := range m {
		err = setColun(f, i, mainSheet, s, strings)

		if err != nil {
			fmt.Println(err)
		}
		i++
	}

	/*	f.SetCellValue(mainSheet, "A1", "Número")
		f.SetCellValue(mainSheet, "B1", "Mês")*/

	// Adiciona os meses na planilha de configuração para a validação de dados
	//	months := []string{"Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho", "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"}

	/*	// Define a validação de dados para a coluna "Mês" (B2:B1000) usando uma lista de strings
		dvMonth := excelize.NewDataValidation(true)
		dvMonth.Sqref = "B2:B100000"
		dvMonth.SetDropList(months)
		if err := f.AddDataValidation(mainSheet, dvMonth); err != nil {
			fmt.Println(err)
			return
		}
	*/
	// Define a validação de dados para a coluna "Ano" (C2:C1000) com anos de 2010 a 2030
	years := []string{}
	for year := 2010; year <= 2030; year++ {
		years = append(years, fmt.Sprintf("%d", year))
	}

	/*	// Define a validação de dados para a coluna "Ano" (C2:C1000) com anos de 2010 a 2030
		dvAno := excelize.NewDataValidation(true)
		dvAno.Sqref = "C2:C100000"
		dvAno.SetDropList(years)
		if err := f.AddDataValidation(mainSheet, dvAno); err != nil {
			fmt.Println(err)
			return
		}*/

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

func setColun[T any](f *excelize.File, indexColun uint, sheet, nameColun string, values []T) error {
	// Preciso pegar o nome da coluna
	cellName := getColumnName(indexColun)
	fmt.Println(fmt.Sprintf("%s1", cellName))
	err := f.SetCellValue(sheet, fmt.Sprintf("%s1", cellName), nameColun)
	if err != nil {
		return err
	}
	if len(values) > 1 {
		dvAno := excelize.NewDataValidation(true)
		dvAno.Sqref = fmt.Sprintf("%s2:%s100000", cellName, cellName)

		sliceString := convertToStringSlice(values)
		dvAno.SetDropList(sliceString)
		if err = f.AddDataValidation(sheet, dvAno); err != nil {
			return err
		}
	}
	return nil

}

func getColumnName(n uint) string {
	if n < 0 || n > 25 {
		return ""
	}
	return columnLetters[n]
}

// Função genérica que converte qualquer slice para um slice de strings
func convertToStringSlice[T any](s []T) []string {
	var result []string
	for _, v := range s {
		// Convertendo o valor para string
		str := fmt.Sprintf("%v", v)
		result = append(result, str)
	}
	return result
}

func structToMap(s interface{}) map[string][]string {
	result := make(map[string][]string)

	// Obtém o valor e tipo do struct
	value := reflect.ValueOf(s)
	typeOfS := reflect.TypeOf(s)

	// Certifica-se de que o valor é um struct
	if value.Kind() != reflect.Struct {
		fmt.Println("Expected a struct")
		return nil
	}

	// Itera sobre os campos do struct
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := typeOfS.Field(i)

		// Obtém o nome do campo e o valor
		fieldName := fieldType.Name
		fieldValue := field.Interface()

		// Converte o valor do campo para um slice de strings
		var fieldValueSlice []string
		switch v := fieldValue.(type) {
		case string:
			fieldValueSlice = []string{v}
		case int:
			fieldValueSlice = []string{strconv.Itoa(v)}
		case float64:
			fieldValueSlice = []string{strconv.FormatFloat(v, 'f', -1, 64)}
		case []string:
			fieldValueSlice = v
		// Adicione mais casos conforme necessário
		default:
			fieldValueSlice = []string{fmt.Sprintf("%v", v)}
		}

		// Adiciona o nome do campo e o valor ao mapa
		result[fieldName] = fieldValueSlice
	}

	return result
}
