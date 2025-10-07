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
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", true),
	)

	term := "golang documentation"

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), options...)
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Tasks{
			chromedp.Navigate("https://www.google.com"),
			chromedp.WaitReady("body", chromedp.ByQuery),

			chromedp.WaitVisible(`textarea[name="q"]`, chromedp.ByQuery),
			chromedp.SendKeys(`textarea[name="q"]`, term, chromedp.ByQuery),

			chromedp.Click(`input[name="btnI"]`, chromedp.ByQuery),

			chromedp.Sleep(25 * time.Second),
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Página 'Estou com sorte' aberta com sucesso!")
}
