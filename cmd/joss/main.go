package main

import (
	"fmt"
	"os"

	"github.com/jossecurity/joss/pkg/server"
	"github.com/jossecurity/joss/pkg/core"
	"github.com/jossecurity/joss/pkg/parser"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "server":
		if len(os.Args) >= 3 && os.Args[2] == "start" {
			server.Start()
		} else {
			fmt.Println("Uso: joss server start")
		}
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Uso: joss run [archivo.joss]")
			return
		}
		filename := os.Args[2]
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Printf("Error leyendo archivo: %v\n", err)
			return
		}
		
		l := parser.NewLexer(string(data))
		p := parser.NewParser(l)
		program := p.ParseProgram()
		
		if len(p.Errors()) != 0 {
			fmt.Println("Errores de parseo:")
			for _, msg := range p.Errors() {
				fmt.Printf("\t%s\n", msg)
			}
			return
		}
		
		rt := core.NewRuntime()
		rt.Execute(program)

	case "version":
		fmt.Println("JosSecurity v3.0 (Gold Master)")
	case "help":
		printHelp()
	default:
		fmt.Printf("Comando desconocido: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Uso: joss [comando] [argumentos]")
	fmt.Println("Comandos disponibles:")
	fmt.Println("  server start   Inicia el servidor HTTP de desarrollo")
	fmt.Println("  version        Muestra la versión actual")
	fmt.Println("  help           Muestra esta ayuda")
}
