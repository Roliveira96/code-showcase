package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Minha Aplicação")

	// Criar a lista
	list := widget.NewList(
		func() int {
			return 10 // Número de itens na lista
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Item") // Widget para cada item
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText("Item ") // Corrigido: índice da lista começa em 1
		},
	)

	// Criar o widget flutuante
	floatingButton := widget.NewButton("Flutuante", func() {
		// Criar a nova janela
		newWindow := a.NewWindow("Janela Flutuante")

		// Impedir que a janela seja minimizada
		newWindow.SetFixedSize(true)
		newWindow.CenterOnScreen()
		newWindow.RequestFocus()
		newWindow.ShowAndRun()

		// Conteúdo da nova janela com cor de fundo transparente
		newWindow.SetContent(canvas.NewRectangle(color.RGBA{123, 123, 123, 122})) // Cor de fundo semi-transparente
		newWindow.Resize(fyne.NewSize(300, 400))
		// Mostrar a nova janela
		newWindow.Show()
	})

	// Conteúdo da janela principal
	w.SetContent(container.New(layout.NewBorderLayout(nil, floatingButton, nil, nil), list, floatingButton))

	// Definir tamanho da janela
	w.Resize(fyne.NewSize(300, 400))

	w.ShowAndRun()
}
