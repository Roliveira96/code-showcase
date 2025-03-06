package main

import (
	"fmt"
	"time"
)

// calcularDataMensal adiciona meses a uma data do tipo time.Time e retorna uma string formatada.
// Se o dia do mês resultante for inválido, ajusta para o último dia do mês.
func calcularDataMensal(date time.Time, months int) string {
	year, month, day := date.Date()
	newMonth := int(month) + months
	newYear := year + newMonth/12
	newMonth = newMonth % 12
	if newMonth == 0 {
		newMonth = 12
		newYear--
	}

	newTime := time.Date(newYear, time.Month(newMonth), day, 0, 0, 0, 0, date.Location())
	if newTime.Day() != day {
		newTime = time.Date(newYear, time.Month(newMonth+1), 1, 0, 0, 0, 0, date.Location()).AddDate(0, 0, -1)
	}

	return newTime.Format("02-01-2006")
}

// formatarData formata a data no formato "02-01-2006".
func formatarData(data time.Time) string {
	return data.Format("02-01-2006")
}

func main() {
	meses := []time.Month{
		time.January, time.February, time.March, time.April,
		time.May, time.June, time.July, time.August,
		time.September, time.October, time.November, time.December,
	}

	for _, mes := range meses {
		for dia := 1; dia <= 31; dia++ {
			// Ignora dias inválidos para o mês
			if dia > diasNoMes(2025, mes) {
				continue
			}

			dataInicial := time.Date(2025, mes, dia, 0, 0, 0, 0, time.UTC)

			for recorrencia := 1; recorrencia <= 12; recorrencia++ {
				dataFinal := calcularDataMensal(dataInicial, recorrencia)
				fmt.Printf("Data inicial: %s, Recorrência: %d, Data final: %s\n", dataInicial.Format("02-01-2006"), recorrencia, dataFinal)
			}
		}
	}
}

// diasNoMes retorna o número de dias em um mês específico.
func diasNoMes(ano int, mes time.Month) int {
	if mes == time.February {
		if (ano%4 == 0 && ano%100 != 0) || ano%400 == 0 {
			return 29 // Ano bissexto
		}
		return 28
	} else if mes == time.April || mes == time.June || mes == time.September || mes == time.November {
		return 30
	} else {
		return 31
	}
}
