package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

func main() {
	for a := uint8(255); a >= 0; a-- {
		for b := uint8(255); b >= 0; b-- {
			for c := uint8(255); c >= 0; c-- {
				for d := uint8(255); d >= 0; d-- {
					ip := fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)

					// Comando ping diferente para Windows e Linux/macOS
					var cmd *exec.Cmd
					if runtime.GOOS == "windows" {
						cmd = exec.Command("ping", "-n", "1", ip)
					} else {
						cmd = exec.Command("ping", "-c", "1", ip)
					}

					// Executa o comando ping
					err := cmd.Run()
					if err == nil {
						fmt.Println(ip, "- Host ativo")
					} else {
						fmt.Println(ip, "- Host inativo")
					}
				}
			}
		}
	}
}
