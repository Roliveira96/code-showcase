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

	log.Println("Imagem salva como radar_chart.png")

	var indicators []*opts.Indicator

	indicators = []*opts.Indicator{
		{Name: "Vendas", Max: 100, Min: 2, Color: "#131313"},
		{Name: "Marketing", Max: 100, Color: "#131313"},
		{Name: "Pesquisa", Max: 100, Color: "#131313"},
		{Name: "Desenvolvimento", Max: 100, Color: "#131313"},
		{Name: "Suporte", Max: 100, Color: "#131313"},
	}

	dataA := []float64{80, 90, 70, 85, 95}
	dataB := []float64{60, 85, 75, 93, 80}
	dataC := []float64{10, 90, 70, 85, 95}
	dataD := []float64{20, 85, 35, 95, 80}

	radar := charts.NewRadar()
	radar.SetGlobalOptions()

	colors := []string{"#ff000080", "#00ff0080", "#0000ff80", "#ff00ff80", "#f1016290"}
	teams := []string{"Equipe A", "Equipe B", "Equipe C", "Equipe D", "Equipe E"} // Nomes das equipes
	data := [][]float64{dataA, dataB, dataA, dataC, dataD}

	radar.SetGlobalOptions(
		charts.WithRadarComponentOpts(opts.RadarComponent{
			Center: []string{"63%", "50%"},
		}),
	)
	radar.RadarComponent.Indicator = indicators

	show := true

	for i, team := range teams {
		radar.AddSeries(team, []opts.RadarData{{Value: data[i], Name: team}}).
			SetSeriesOptions(
				charts.WithLineStyleOpts(opts.LineStyle{Width: 1}),
				charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: 0.2}),
				charts.WithItemStyleOpts(opts.ItemStyle{Color: colors[i]}),
				charts.WithLabelOpts(opts.Label{Show: &show, Position: "left"}),
			)
	}

	f, _ := os.Create("radar_chart.html")
	defer f.Close()
	radar.Render(f)

	err := htmlToImage("radar_chart.html", "radar_chart.png", 500, 500)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Imagem salva como radar_chart.png")
}

func htmlToImage(htmlPath, imgPath string, width, height int) error {

	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", true),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), options...)
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	absolutePath, err := filepath.Abs(htmlPath)
	if err != nil {
		return err
	}
	fileURL := "file://" + absolutePath

	var buf []byte
	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.EmulateViewport(int64(width), int64(height)),
		chromedp.Navigate(fileURL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(1 * time.Second),
		chromedp.FullScreenshot(&buf, 100),
	})
	if err != nil {
		return err
	}

	err = os.WriteFile(imgPath, buf, 0644)
	if err != nil {
		return err
	}

	return nil
}
