package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func main() {

	log.Println("Imagem salva como radar_chart.png") // Image saved as radar_chart.png

	// Dados fictícios para o gráfico de radar / Fictional data for the radar chart
	var indicators []*opts.Indicator

	indicators = []*opts.Indicator{
		{Name: "Vendas", Max: 100, Min: 2, Color: "#131313"},  // Sales (Red with 50% transparency)
		{Name: "Marketing", Max: 100, Color: "#131313"},       // Marketing (Blue with 50% transparency)
		{Name: "Pesquisa", Max: 100, Color: "#131313"},        // Research (Black with 50% transparency)
		{Name: "Desenvolvimento", Max: 100, Color: "#131313"}, // Development
		{Name: "Suporte", Max: 100, Color: "#131313"},         // Support
	}

	dataA := []float64{80, 90, 70, 85, 95}
	dataB := []float64{60, 85, 75, 93, 80}
	dataC := []float64{10, 90, 70, 85, 95}
	dataD := []float64{20, 85, 35, 95, 80}

	// Criar o gráfico de radar / Create the radar chart
	radar := charts.NewRadar()
	radar.SetGlobalOptions()

	teste := false
	// Adicionar séries com cores diferentes
	// Definir cores e equipes
	colors := []string{"#ff000080", "#00ff0080", "#0000ff80", "#ff00ff80", "#f1016290"}
	teams := []string{"Equipe A", "Equipe B", "Equipe C", "Equipe D", "Equipe E"} // Nomes das equipes
	data := [][]float64{dataA, dataB, dataA, dataC, dataD}

	// Configurar o gráfico para alinhar à esquerda
	radar.SetGlobalOptions(
		charts.WithRadarComponentOpts(opts.RadarComponent{
			Center: []string{"63%", "50%"}, // Ajuste a posição conforme necessário
		}),
	)
	radar.RadarComponent.Indicator = indicators

	// Adicionar séries com cores diferentes
	for i, team := range teams {
		radar.AddSeries(team, []opts.RadarData{{Value: data[i], Name: team}}).
			SetSeriesOptions(
				charts.WithLineStyleOpts(opts.LineStyle{Width: 1}),               // Largura da linha
				charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: 0.2}),           // Opacidade da área
				charts.WithItemStyleOpts(opts.ItemStyle{Color: colors[i]}),       // Cor dos itens
				charts.WithLabelOpts(opts.Label{Show: &teste, Position: "left"}), // Mostrar rótulos
			)
	}

	// Renderizar o gráfico em um arquivo HTML / Render the chart to an HTML file
	f, _ := os.Create("radar_chart.html")
	defer f.Close()
	radar.Render(f)

	err := htmlToImage("radar_chart.html", "radar_chart.png", 500, 500)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Imagem salva como radar_chart.png") // Image saved as radar_chart.png
}

// Função para converter HTML em imagem
func htmlToImage(htmlPath, imgPath string, width, height int) error {
	// Configurar o contexto do Chrome
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // Executar o Chrome em modo não headless
		chromedp.Flag("disable-gpu", true),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), options...)
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Converter o caminho do arquivo para o formato de URL
	absolutePath, err := filepath.Abs(htmlPath)
	if err != nil {
		return err
	}
	fileURL := "file://" + absolutePath

	// Inicializar a captura de imagem
	var buf []byte
	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.EmulateViewport(int64(width), int64(height)), // Define a largura e altura da viewport
		chromedp.Navigate(fileURL),                            // Navegar diretamente para o arquivo HTML
		chromedp.WaitVisible("body", chromedp.ByQuery),        // Aguarda o corpo da página estar visível
		chromedp.Sleep(1 * time.Second),                       // Espera para garantir que tudo foi carregado
		chromedp.FullScreenshot(&buf, 100),                    // Captura uma screenshot da página
	})
	if err != nil {
		return err
	}

	// Salvar o buffer de bytes como um arquivo PNG
	err = os.WriteFile(imgPath, buf, 0644)
	if err != nil {
		return err
	}

	return nil
}
