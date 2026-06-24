package main

import (
	"fmt"
	"strings"
	"time"
)

type Grupo interface {
	AnalisarSociedade()
}

type Individuo struct {
	Posicao    string
	Patrimonio float64
}

type CadeiaHereditaria struct {
	Membros []Individuo
}

func (c CadeiaHereditaria) AnalisarSociedade() {
	for _, m := range c.Membros {
		fmt.Printf("Analisando grupo... ")

		if strings.ToLower(m.Posicao) == "cima" {
			m.Patrimonio *= 1.5
			fmt.Println("O de cima sobe, rico cada vez mais rico.")
		} else {
			m.Patrimonio *= 0.5
			fmt.Println("O de baixo desce, pobre cada vez mais pobre.")
		}

		time.Sleep(1 * time.Second)
	}

	fmt.Println("Bom, xibom, bombom...")
	c.AnalisarSociedade()
}

func main() {
	populacao := CadeiaHereditaria{
		Membros: []Individuo{
			{Posicao: "cima", Patrimonio: 1000000},
			{Posicao: "baixo", Patrimonio: 100},
		},
	}

	populacao.AnalisarSociedade()
}
