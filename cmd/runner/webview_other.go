//go:build !windows

package main

import (
	"log"
)

func runGUIOrWait(finalPort string) {
	log.Println("[Joss Runner] Modo Servidor/Headless (Linux / macOS / Non-Windows).")
	log.Printf("[Joss Runner] Servidor activo escuchando en http://localhost:%s\n", finalPort)
	waitForSignal()
}
