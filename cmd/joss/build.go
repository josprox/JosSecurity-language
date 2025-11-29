package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

func buildWeb() {
	fmt.Println("Iniciando compilación WEB de JosSecurity...")

	// 1. Validate Structure (Strict Topology)
	required := []string{
		"main.joss",
		"env.joss",
		"app",
		"config",
		"api.joss",
		"routes.joss",
	}
	for _, f := range required {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			fmt.Printf("Error de Arquitectura: Falta archivo/directorio requerido '%s'\n", f)
			fmt.Println("La Biblia de JosSecurity requiere una estructura estricta.")
			return
		}
	}

	// 2. Encrypt env.joss
	fmt.Println("Encriptando entorno...")
	encryptEnvTo("env.enc")

	fmt.Println("Build WEB completado exitosamente.")
}

func buildProgram() {
	fmt.Println("Iniciando compilación PROGRAM de JosSecurity...")

	// 1. Ask for Target OS
	fmt.Println("Seleccione el sistema operativo destino:")
	fmt.Println("1. Windows")
	fmt.Println("2. Linux")
	fmt.Println("3. MacOS")
	fmt.Print("Opción: ")

	var option string
	fmt.Scanln(&option)

	goos := "windows"
	ext := ".exe"

	switch option {
	case "1", "windows":
		goos = "windows"
		ext = ".exe"
	case "2", "linux":
		goos = "linux"
		ext = ""
	case "3", "macos", "darwin":
		goos = "darwin"
		ext = ""
	default:
		fmt.Println("Opción inválida. Usando Windows por defecto.")
	}

	fmt.Printf("Compilando para %s...\n", goos)

	// 2. Build Binary
	// We need to run 'go build' with env vars.
	// We are compiling the current package (cmd/joss).
	cmd := exec.Command("go", "build", "-o", "dist/joss_app"+ext, "./cmd/joss")
	cmd.Env = append(os.Environ(), "GOOS="+goos, "CGO_ENABLED=0") // Disable CGO for easier cross-compile

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error compilando: %v\n%s\n", err, output)
		return
	}

	// 3. Copy Assets
	fmt.Println("Copiando archivos del proyecto...")
	copyDir("app", "dist/app")
	copyDir("config", "dist/config")
	copyDir("assets", "dist/assets")
	copyFile("main.joss", "dist/main.joss")
	copyFile("api.joss", "dist/api.joss")
	copyFile("routes.joss", "dist/routes.joss")

	// Encrypt Env
	encryptEnvTo("dist/env.enc")

	fmt.Println("Build PROGRAM completado en carpeta 'dist/'.")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		return copyFile(path, dstPath)
	})
}

func encryptEnvTo(destPath string) {
	data, err := ioutil.ReadFile("env.joss")
	if err != nil {
		fmt.Println("Error leyendo env.joss")
		return
	}

	key := []byte("12345678901234567890123456789012")
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("Error cipher:", err)
		return
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Println("Error GCM:", err)
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println("Error nonce:", err)
		return
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	err = ioutil.WriteFile(destPath, ciphertext, 0644)
	if err != nil {
		fmt.Println("Error escribiendo env.enc")
		return
	}
}
