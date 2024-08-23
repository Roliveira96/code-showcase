package main

import (
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func main() {
	// Dados fictícios para o gráfico de radar / Fictional data for the radar chart
	indicators := []*opts.Indicator{
		{Name: "Vendas", Max: 100},          // Sales
		{Name: "Marketing", Max: 100},       // Marketing
		{Name: "Pesquisa", Max: 100},        // Research
		{Name: "Desenvolvimento", Max: 100}, // Development
		{Name: "Suporte", Max: 100},         // Support
	}

	data := []opts.RadarData{
		{
			Name:  "Equipe A", // Team A
			Value: []float64{80, 90, 70, 85, 95},
		},
		{
			Name:  "Equipe B", // Team B
			Value: []float64{60, 85, 75, 95, 80},
		},
	}

	// Criar o gráfico de radar / Create the radar chart
	radar := charts.NewRadar()
	radar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Gráfico de Radar", // Radar Chart
		}),
		charts.WithRadarComponentOpts(opts.RadarComponent{
			Indicator: indicators,
		}),
	)

	radar.AddSeries("Radar", data).
		SetSeriesOptions(
			charts.WithLineStyleOpts(opts.LineStyle{Color: "red"}), // Cor da linha / Line color
			charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: 0.2}), // Opacidade da área / Area opacity
		)

	// Salvar o gráfico em um arquivo PNG / Save the chart as a PNG file
	f, _ := os.Create("radar_chart.html") // Criar um arquivo HTML / Create an HTML file
	defer f.Close()
	radar.Render(f) // Renderizar o gráfico / Render the chart

	// Observação: Para salvar como PNG, você precisa de uma ferramenta adicional como o `wkhtmltoimage` ou usar uma biblioteca que suporte exportação de imagem diretamente.
	// Note: To save as PNG, you'll need an additional tool like `wkhtmltoimage` or use a library that supports direct image export.
}
