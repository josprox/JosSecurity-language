package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func handleBrevoConfig() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("--- Configuración de Brevo API ---")
	fmt.Print("¿Deseas activar BREVO_API? (y/n): ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	envPath := "env.joss"
	// Check if env.joss exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		fmt.Println("Error: No se encontró el archivo env.joss en el directorio actual.")
		return
	}

	if response == "y" || response == "yes" || response == "s" || response == "si" {
		fmt.Print("Introduce tu Brevo API Key: ")
		key, _ := reader.ReadString('\n')
		key = strings.TrimSpace(key)

		if key == "" {
			fmt.Println("Error: La API Key no puede estar vacía.")
			return
		}

		updateEnvFile(envPath, "BREVO_API", key)
		fmt.Println("✅ BREVO_API activado y configurado exitosamente.")

	} else {
		// Disable (Comment out or remove)
		removeEnvKey(envPath, "BREVO_API")
		fmt.Println("✅ BREVO_API desactivado.")
	}
}

// Helper to remove or comment out a key
func removeEnvKey(path, key string) {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error leyendo %s: %v\n", path, err)
		return
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	for _, line := range lines {
		// Remove lines starting with key=
		if strings.HasPrefix(strings.TrimSpace(line), key+"=") {
			continue
		}
		newLines = append(newLines, line)
	}

	os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644)
}
