package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
)

func main() {
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	func1()
	func2()
	func3()
	func4()
	func5()

	fmt.Println("Execute `go tool pprof cpu.prof` para analisar")
	fmt.Println("  - Use `top10` para ver as 10 funções que mais consomem CPU.")
	fmt.Println("  - Use `list func3` para ver o tempo gasto em cada linha da função func3.")
	fmt.Println("  - Use `web` para gerar um gráfico interativo no navegador.")
	fmt.Println("  - Use `pdf` para gerar um gráfico em PDF.")
	fmt.Println("  - Use `tree` para visualizar a árvore de chamadas.")
	fmt.Println("  - Use `disasm func3` para ver o código assembly da função func3.")
	fmt.Println("Execute `go tool pprof -http=:8080 cpu.prof` para analisar o gráfico no navegador")
	fmt.Println("Execute `go tool pprof --focus=func3 cpu.prof` para focar na função func3.")
	fmt.Println("Execute `go tool pprof --ignore=fmt cpu.prof` para ignorar funções do pacote fmt na análise.")
	fmt.Println("Para comparar dois perfis, gere outro arquivo (ex: cpu2.prof) e execute:")
	fmt.Println("  `go tool pprof --base cpu.prof cpu2.prof`")
}

func func1() {
	for i := 0; i < 1000000; i++ {
		_ = rand.Intn(1000)
	}
}

func func2() {
	for i := 0; i < 500000; i++ {
		_ = i * i
	}
}

func func3() {
	var s string
	for i := 0; i < 100000; i++ {
		s += "a"
	}
}

func func4() {
	for i := 0; i < 50000; i++ {
		_ = fmt.Sprintf("%d", i)
	}
}

func func5() {
	for i := 0; i < 10000; i++ {
		_ = i * i * i
	}
}
