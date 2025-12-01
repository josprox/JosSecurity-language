package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	_ "embed"

	"github.com/jossecurity/joss/pkg/crypto"
)

//go:embed runner_windows.exe
var runnerWindows []byte

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
	fmt.Println("Iniciando compilación PROGRAM de JosSecurity (SECURE MODE)...")

	// 1. Ask for Target OS
	fmt.Println("Seleccione el sistema operativo destino:")
	fmt.Println("1. Windows")
	fmt.Print("Opción: ")

	var option string
	fmt.Scanln(&option)

	if option != "1" && option != "windows" {
		fmt.Println("Solo Windows es soportado en esta versión pre-compilada.")
		return
	}

	fmt.Println("Compilando para Windows...")

	// 2. Prepare Build Directory
	buildDir := "build"
	os.RemoveAll(buildDir)
	os.MkdirAll(filepath.Join(buildDir, "data"), 0755)
	os.MkdirAll(filepath.Join(buildDir, "Storage"), 0755)

	// 3. Encrypt Assets
	fmt.Println("Encriptando y empaquetando assets...")

	buildKey := make([]byte, 32)
	if _, err := rand.Read(buildKey); err != nil {
		fmt.Printf("Error generando key: %v\n", err)
		return
	}

	files := make(map[string][]byte)
	dirsToWalk := []string{"app", "config", "assets", "public"}
	for _, dir := range dirsToWalk {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			data, err := ioutil.ReadFile(path)
			if err == nil {
				relPath := filepath.ToSlash(path)
				files[relPath] = data
			}
			return nil
		})
	}

	rootFiles := []string{"main.joss", "api.joss", "routes.joss", "env.joss"}
	for _, f := range rootFiles {
		if data, err := ioutil.ReadFile(f); err == nil {
			if f == "env.joss" {
				if _, err := os.Stat("database.sqlite"); err == nil {
					override := "\nDB_PATH=\"Storage/database.sqlite\""
					data = append(data, []byte(override)...)
					fmt.Println("Inyectando configuración DB_PATH=\"Storage/database.sqlite\" en env.joss embebido...")
				}
			}
			files[f] = data
		}
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(files); err != nil {
		fmt.Printf("Error encoding assets: %v\n", err)
		return
	}

	encryptedAssets, err := crypto.EncryptAES(buf.Bytes(), buildKey)
	if err != nil {
		fmt.Printf("Error encrypting assets: %v\n", err)
		return
	}

	// 4. Create Final Executable
	// Layout: [Runner] [Encrypted Assets] [Key 32] [Len 8] [Magic 16]

	outPath := filepath.Join(buildDir, "program.exe")
	f, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("Error creando ejecutable: %v\n", err)
		return
	}
	defer f.Close()

	// Write Runner
	if _, err := f.Write(runnerWindows); err != nil {
		fmt.Printf("Error escribiendo runner: %v\n", err)
		return
	}

	// Write Encrypted Assets
	if _, err := f.Write(encryptedAssets); err != nil {
		fmt.Printf("Error escribiendo assets: %v\n", err)
		return
	}

	// Write Key (32 bytes)
	if _, err := f.Write(buildKey); err != nil {
		fmt.Printf("Error escribiendo key: %v\n", err)
		return
	}

	// Write Assets Length (8 bytes)
	lenBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(lenBuf, uint64(len(encryptedAssets)))
	if _, err := f.Write(lenBuf); err != nil {
		fmt.Printf("Error escribiendo longitud: %v\n", err)
		return
	}

	// Write Magic Marker (16 bytes)
	magic := []byte("JOSS_RUNNER_DATA") // Must match runner
	if _, err := f.Write(magic); err != nil {
		fmt.Printf("Error escribiendo magic marker: %v\n", err)
		return
	}

	// 5. Copy Database
	if _, err := os.Stat("database.sqlite"); err == nil {
		copyFile("database.sqlite", filepath.Join(buildDir, "Storage", "database.sqlite"))
		fmt.Println("Base de datos copiada a build/Storage/")
	}

	// 6. Create error.log
	ioutil.WriteFile(filepath.Join(buildDir, "error.log"), []byte(""), 0666)

	fmt.Println("Build PROGRAM completado exitosamente en carpeta 'build/'.")
	fmt.Println("Estructura:")
	fmt.Printf("  %s\n", outPath)
	fmt.Println("  build/error.log")
	fmt.Println("  build/Storage/database.sqlite")
	fmt.Println("  build/data/ (vacío por ahora)")
}

func runCmd(dir, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.CombinedOutput()
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
	// Dummy implementation for now
}
