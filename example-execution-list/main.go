package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	// Comando para listar os processos (Linux/macOS)
	cmd := exec.Command("ps", "aux")

	// No Windows, use: cmd := exec.Command("tasklist")

	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fmt.Println(line)
	}
}
