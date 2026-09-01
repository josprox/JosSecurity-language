package main

import (
	"fmt"
	"os"

	"github.com/jossecurity/joss/pkg/formatter"
)

func handleFormatCommand(args []string) {
	write := false
	check := false
	var targetPath string

	for _, arg := range args {
		switch arg {
		case "--write", "-w":
			write = true
		case "--check", "-c":
			check = true
		default:
			if targetPath == "" {
				targetPath = arg
			}
		}
	}

	if targetPath == "" {
		targetPath = "."
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		fmt.Printf("Error: No se pudo acceder a '%s': %v\n", targetPath, err)
		os.Exit(1)
	}

	if !info.IsDir() {
		changed, err := formatter.FormatFile(targetPath, write || !check)
		if err != nil {
			fmt.Printf("Error formateando %s: %v\n", targetPath, err)
			os.Exit(1)
		}
		if check {
			if changed {
				fmt.Printf("[FORMAT ERROR] %s no cumple con el formato canónico de Joss.\n", targetPath)
				os.Exit(1)
			}
			fmt.Printf("[FORMAT OK] %s está correctamente formateado.\n", targetPath)
			return
		}
		if changed {
			fmt.Printf("[FORMAT] %s formateado correctamente.\n", targetPath)
		} else {
			fmt.Printf("[FORMAT] %s ya está en formato canónico.\n", targetPath)
		}
		return
	}

	unformatted, err := formatter.FormatDirectory(targetPath, write, check)
	if err != nil {
		fmt.Printf("Error procesando directorio %s: %v\n", targetPath, err)
		os.Exit(1)
	}

	if check {
		if len(unformatted) > 0 {
			fmt.Printf("[FORMAT ERROR] Se encontraron %d archivo(s) sin formatear:\n", len(unformatted))
			for _, file := range unformatted {
				fmt.Printf("  - %s\n", file)
			}
			os.Exit(1)
		}
		fmt.Println("[FORMAT OK] Todos los archivos .joss cumplen con el formato canónico.")
		return
	}

	if write {
		fmt.Printf("[FORMAT] %d archivo(s) modificados y formateados.\n", len(unformatted))
	} else {
		if len(unformatted) > 0 {
			fmt.Printf("[FORMAT] %d archivo(s) requieren formato (usa --write para aplicar):\n", len(unformatted))
			for _, file := range unformatted {
				fmt.Printf("  - %s\n", file)
			}
		} else {
			fmt.Println("[FORMAT] Todos los archivos .joss están correctamente formateados.")
		}
	}
}
