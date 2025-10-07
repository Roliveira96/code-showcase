package main

import (
	"fmt"
	"time"

	"github.com/go-vgo/robotgo"
)

func main() {
	lastActivity := time.Now()
	var xold, yold int
	for {
		// Verifica a posição do mouse a cada 5 segundos
		x, y := robotgo.Location()

		fmt.Println(x, y)
		// Se a posição do mouse mudou, o usuário está ativo
		if x != xold || y != yold {
			lastActivity = time.Now()
			fmt.Println("Usuário ativo!")
		}

		// Verifica a ociosidade
		if time.Since(lastActivity) > 7*time.Second {
			fmt.Println("Usuário ausente!")
		}

		xold, yold = x, y

		time.Sleep(5 * time.Second)
	}
}
