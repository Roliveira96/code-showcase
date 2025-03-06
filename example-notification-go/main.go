package main

import (
	"fmt"
	_ "image/color"
	"log"
	"time"

	"github.com/fogleman/gg"
)

func main() {
	// Cria uma goroutine para controlar a exibição do retângulo
	go func() {
		for {
			// Desenha o retângulo
			desenharRetangulo()

			// Aguarda 10 segundos
			time.Sleep(10 * time.Second)

			// Limpa a tela (ou redesenha a tela sem o retângulo)
			limparTela()

			// Aguarda 30 segundos
			time.Sleep(30 * time.Second)
		}
	}()

	// Aguarda indefinidamente
	fmt.Scanln()
}

func desenharRetangulo() {
	const S = 1024
	dc := gg.NewContext(S, S)
	dc.SetRGBA(0, 0, 0, 0.5) // Define a cor com transparência (alfa = 0.5)
	dc.DrawRectangle(100, 100, 400, 400)
	dc.Fill()
	dc.SetRGB(1, 1, 1)                                      // Define a cor do texto
	dc.DrawStringAnchored("Beba água!", S/2, S/2, 0.5, 0.5) // Escreve o texto no centro do retângulo
	err := dc.SavePNG("retangulo.png")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Desenho salvo em retangulo.png")
}

func limparTela() {
	// Implemente a lógica para limpar a tela aqui
	// Pode ser algo como redesenhar a tela sem o retângulo,
	// ou usar uma função específica da sua biblioteca gráfica
}
