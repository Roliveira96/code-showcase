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

	if err = f.DeleteSheet("Sheet1"); err != nil {
		fmt.Println("Erro ao remover a planilha:", err)
		return
	}

	if err = f.SaveAs("example.xlsx"); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Arquivo Excel "
	example.xlsx
	" criado com sucesso.")
}

func setColun[T any](f *excelize.File, indexColun uint, sheet, nameColun string, values []T) error {
	letterCell := getColumnName(indexColun)

	err := f.SetCellValue(sheet, fmt.Sprintf("%s1", letterCell), nameColun)
	if err != nil {
		return err
	}
	if len(values) > 1 {
		dv := excelize.NewDataValidation(true)
		dv.Sqref = fmt.Sprintf("%s2:%s100000", letterCell, letterCell)

		sliceString := convertToStringSlice(values)

		if err = dv.SetDropList(sliceString); err != nil {
			return err
		}

		if err = f.AddDataValidation(sheet, dv); err != nil {
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
		str := fmt.Sprintf("%v", v)
		result = append(result, str)
	}
	return result
}

func structToMap(s interface{}) map[string][]string {
	result := make(map[string][]string)

	value := reflect.ValueOf(s)
	typeOfS := reflect.TypeOf(s)

	if value.Kind() != reflect.Struct {
		fmt.Println("Expected a struct")
		return nil
	}

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := typeOfS.Field(i)

		fieldName := fieldType.Name
		fieldValue := field.Interface()

		var fieldValueSlice []string
		switch v := fieldValue.(type) {
		case string:
			fieldValueSlice = []string{v}
		case int:
			fieldValueSlice = []string{strconv.Itoa(v)}
		case float64:
			fieldValueSlice = []string{strconv.FormatFloat(v, "f", -1, 64)}
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
