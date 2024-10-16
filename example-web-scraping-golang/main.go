package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // Executar o Chrome em modo não headless
		chromedp.Flag("disable-gpu", true),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), options...)
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate("https://www.google.com"),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(5 * time.Second),

		// Clica em "Imagens"
		chromedp.WaitReady(`a.gb_V[aria-label="Pesquisar imagens "]`, chromedp.ByQuery),
		chromedp.Click(`a.gb_V[aria-label="Pesquisar imagens "]`, chromedp.ByQuery),

		// Espera a página de imagens carregar
		chromedp.WaitReady("#islrg", chromedp.ByID),
		chromedp.Sleep(5 * time.Second),

		// Preenche o campo de pesquisa e envia a pesquisa (usando Focus e ActionFunc)
		chromedp.WaitReady(`textarea[name="q"]`, chromedp.ByQuery),
		chromedp.Focus(`textarea[name="q"]`, chromedp.ByQuery),
		chromedp.SendKeys(`textarea[name="q"]`, "golang", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.KeyEvent("Enter").Do(ctx)
		}),
	})

	if err != nil {
		if err == context.DeadlineExceeded {
			log.Fatal("Timeout: A página demorou muito para carregar ou o elemento não foi encontrado.")
		} else {
			log.Fatal("Erro ao executar as ações no Chrome:", err)
		}
	}

	fmt.Println("Clicado em "
	Imagens
	" com sucesso!")
}
