package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func main() {
	// Dados fictícios para o gráfico de radar / Fictional data for the radar chart
	var indicators []*opts.Indicator

	indicators = []*opts.Indicator{
		{Name: "Vendas", Max: 100, Min: 2, Color: "red"}, // Sales
		{Name: "Marketing", Max: 100},                    // Marketing
		{Name: "Pesquisa", Max: 100},                     // Research
		{Name: "Desenvolvimento", Max: 100},              // Development
		{Name: "Suporte", Max: 100},                      // Support
	}

	dataA := []float64{80, 90, 70, 85, 95}
	dataB := []float64{60, 85, 75, 95, 80}

	// Criar o gráfico de radar / Create the radar chart
	radar := charts.NewRadar()
	radar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Gráfico de Radar", // Radar Chart
		}),
	)

	radar.AddSeries("Equipe A", []opts.RadarData{{Value: dataA, Name: "Equipe A"}}).
		AddSeries("Equipe B", []opts.RadarData{{Value: dataB, Name: "Equipe B"}}).
		SetSeriesOptions(
			charts.WithLineStyleOpts(opts.LineStyle{Width: 2}),           // Largura da linha / Line width
			charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: 0.2}),       // Opacidade da área / Area opacity
			charts.WithItemStyleOpts(opts.ItemStyle{Color: "#f00"}),      // Cor dos itens / Items color
			charts.WithLabelOpts(opts.Label{Show: nil, Position: "top"}), // Mostrar rótulos / Show labels
		)

	radar.RadarComponent.Indicator = indicators

	// Renderizar o gráfico em um arquivo HTML / Render the chart to an HTML file
	f, _ := os.Create("radar_chart.html")
	defer f.Close()
	radar.Render(f)

	// Ler o HTML para uma string / Read the HTML into a string
	htmlData, err := ioutil.ReadFile("radar_chart.html")
	if err != nil {
		log.Fatal(err)
	}

	// Usar Chromedp para renderizar a imagem / Use Chromedp to render the image
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf []byte
	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate("data:text/html," + string(htmlData)),
		chromedp.Sleep(130 * time.Second), // Esperar o carregamento / Wait for the loading
		chromedp.FullScreenshot(&buf, 100),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Salvar o buffer de bytes como um arquivo PNG / Save the byte buffer as a PNG file
	if err := ioutil.WriteFile("radar_chart.png", buf, 0644); err != nil {
		log.Fatal(err)
	}

	log.Println("Imagem salva como radar_chart.png") // Image saved as radar_chart.png
}
