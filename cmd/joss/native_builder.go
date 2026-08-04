package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jossecurity/joss/pkg/bytecode"
	"github.com/jossecurity/joss/pkg/crypto"
	"github.com/jossecurity/joss/pkg/parser"
)

// Supported OS and Architecture targets
var supportedTargets = map[string][]string{
	"windows": {"amd64", "arm64", "386"},
	"linux":   {"amd64", "arm64", "386"},
	"darwin":  {"amd64", "arm64"},
}

// buildNative builds a fully-optimized, standalone native executable binary for the specified OS and Architecture.
func buildNative(targetOS, targetArch string) {
	if targetOS == "" {
		targetOS = runtime.GOOS
	}
	if targetArch == "" {
		targetArch = runtime.GOARCH
	}

	targetOS = strings.ToLower(targetOS)
	targetArch = strings.ToLower(targetArch)

	// Validate target
	archs, osValid := supportedTargets[targetOS]
	if !osValid {
		fmt.Printf("Error: Sistema operativo '%s' no soportado. Opciones: windows, linux, darwin.\n", targetOS)
		return
	}
	archValid := false
	for _, a := range archs {
		if a == targetArch {
			archValid = true
			break
		}
	}
	if !archValid {
		fmt.Printf("Error: Arquitectura '%s' no soportada para %s. Opciones: %s\n", targetArch, targetOS, strings.Join(archs, ", "))
		return
	}

	fmt.Printf("\n=======================================================\n")
	fmt.Printf("🚀 COMPILADOR NATIVO DE JOSS (AOT & Standalone Mode)\n")
	fmt.Printf(" Target OS   : %s\n", strings.ToUpper(targetOS))
	fmt.Printf(" Target Arch : %s\n", strings.ToUpper(targetArch))
	fmt.Printf(" Stripped    : -ldflags=\"-s -w\"\n")
	fmt.Printf(" CGO Enabled : 0 (Enlazado Estático Puro)\n")
	fmt.Printf("=======================================================\n\n")

	// 1. Verify Go toolchain
	if _, err := exec.LookPath("go"); err != nil {
		fmt.Println("Error: No se encontró la herramienta 'go' instalada en el sistema.")
		fmt.Println("Para compilar binarios nativos multiplataforma se requiere Go instalada.")
		return
	}

	// 2. Prepare build directory
	buildDir := "build"
	os.RemoveAll(buildDir)
	if err := os.MkdirAll(filepath.Join(buildDir, "Storage"), 0755); err != nil {
		fmt.Printf("Error creando directorio build: %v\n", err)
		return
	}

	// 3. Package & Encrypt Assets (AOT Bytecode compilation)
	fmt.Println("📦 Empaquetando y precompilando assets (AOT Bytecode AST)...")

	buildKey := make([]byte, 32)
	if _, err := rand.Read(buildKey); err != nil {
		fmt.Printf("Error generando clave criptográfica: %v\n", err)
		return
	}

	files := make(map[string][]byte)
	ignoredDirs := map[string]bool{
		".git":         true,
		".vscode":      true,
		".idea":        true,
		"build":        true,
		"vendor":       true,
		"node_modules": true,
		".gemini":      true,
		".codex":       true,
		".agents":      true,
		".github":      true,
		"dist":         true,
	}

	compiledCount := 0
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || path == "." {
			return nil
		}

		parts := strings.Split(path, string(os.PathSeparator))
		if len(parts) > 0 && ignoredDirs[parts[0]] {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		name := strings.ToLower(info.Name())
		if strings.HasSuffix(name, ".exe") || strings.HasSuffix(name, ".exe~") || strings.HasSuffix(name, ".log") || strings.HasSuffix(name, ".enc") || strings.HasSuffix(name, ".dll") || strings.HasSuffix(name, ".tmp") || strings.HasSuffix(name, ".so") || strings.HasSuffix(name, ".dylib") || name == "runner" || name == "joss" {
			return nil
		}

		// Skip large binary files (> 5MB)
		if info.Size() > 5<<20 {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath := filepath.ToSlash(path)

		// Pre-compile .joss files into Bytecode JP v2 if valid
		if strings.HasSuffix(relPath, ".joss") {
			l := parser.NewLexer(string(data))
			p := parser.NewParser(l)
			prog := p.ParseProgram()
			if len(p.Errors()) == 0 && prog != nil {
				if bc, bcErr := bytecode.Encode(prog); bcErr == nil {
					data = bc
					compiledCount++
				}
			}
		}

		files[relPath] = data
		return nil
	})

	if err != nil {
		fmt.Printf("Error leyendo archivos del proyecto: %v\n", err)
		return
	}

	fmt.Printf("⚡ Pre-compilados %d archivos Joss a bytecode nativo.\n", compiledCount)

	// Handle env.joss or .env
	envPath := "env.joss"
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		if _, err := os.Stat(".env"); err == nil {
			envPath = ".env"
		}
	}

	if data, err := ioutil.ReadFile(envPath); err == nil {
		if _, err := os.Stat("database.sqlite"); err == nil {
			override := "\nDB_PATH=\"Storage/database.sqlite\""
			data = append(data, []byte(override)...)
		}

		salt := make([]byte, 16)
		rand.Read(salt)
		masterSecret := []byte("JOSSECURITY_MASTER_SECRET_2025")
		key := crypto.DeriveKey(masterSecret, salt)
		encrypted, err := crypto.EncryptAES(data, key)
		if err == nil {
			files["env.enc"] = append(salt, encrypted...)
		}
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(files); err != nil {
		fmt.Printf("Error codificando VFS: %v\n", err)
		return
	}

	encryptedAssets, err := crypto.EncryptAES(buf.Bytes(), buildKey)
	if err != nil {
		fmt.Printf("Error encriptando VFS payload: %v\n", err)
		return
	}

	// 4. Cross-compile runner binary via Go toolchain
	fmt.Println("🔨 Compilando ejecutable nativo con la toolchain de Go...")
	tempRunnerDir, err := ioutil.TempDir("", "joss-build-*")
	if err != nil {
		fmt.Printf("Error creando directorio temporal: %v\n", err)
		return
	}
	defer os.RemoveAll(tempRunnerDir)

	tempRunnerBin := filepath.Join(tempRunnerDir, "runner.tmp")

	// Determine runner package path
	runnerPkg := "github.com/jossecurity/joss/cmd/runner"
	if _, err := os.Stat("cmd/runner"); err == nil {
		runnerPkg = "./cmd/runner"
	}

	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", tempRunnerBin, runnerPkg)
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		"GOOS="+targetOS,
		"GOARCH="+targetArch,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error durante compilación Go: %v\nSalida: %s\n", err, string(out))
		return
	}

	runnerBytes, err := ioutil.ReadFile(tempRunnerBin)
	if err != nil {
		fmt.Printf("Error leyendo binario base: %v\n", err)
		return
	}

	// 5. Construct Final Native Executable
	// Layout: [Runner Executable Binary] [Encrypted Assets Payload] [Key 32] [Len 8] [Magic 16]
	projectName := filepath.Base(getWorkingDir())
	if projectName == "." || projectName == "/" || projectName == "" {
		projectName = "app"
	}

	exeName := projectName
	if targetOS == "windows" {
		exeName += ".exe"
	}

	outPath := filepath.Join(buildDir, exeName)
	outFile, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("Error creando ejecutable final: %v\n", err)
		return
	}
	defer outFile.Close()

	// Write base runner binary
	if _, err := outFile.Write(runnerBytes); err != nil {
		fmt.Printf("Error escribiendo base runner: %v\n", err)
		return
	}

	// Write encrypted assets
	if _, err := outFile.Write(encryptedAssets); err != nil {
		fmt.Printf("Error escribiendo assets: %v\n", err)
		return
	}

	// Write key (32 bytes)
	if _, err := outFile.Write(buildKey); err != nil {
		fmt.Printf("Error escribiendo clave: %v\n", err)
		return
	}

	// Write assets length (8 bytes uint64 little endian)
	lenBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(lenBuf, uint64(len(encryptedAssets)))
	if _, err := outFile.Write(lenBuf); err != nil {
		fmt.Printf("Error escribiendo longitud payload: %v\n", err)
		return
	}

	// Write magic marker (16 bytes)
	magic := []byte("JOSS_RUNNER_DATA")
	if _, err := outFile.Write(magic); err != nil {
		fmt.Printf("Error escribiendo marca mágica: %v\n", err)
		return
	}

	// 6. Copy Database if exists
	if _, err := os.Stat("database.sqlite"); err == nil {
		copyFile("database.sqlite", filepath.Join(buildDir, "Storage", "database.sqlite"))
		if _, err := os.Stat("database.sqlite-shm"); err == nil {
			copyFile("database.sqlite-shm", filepath.Join(buildDir, "Storage", "database.sqlite-shm"))
		}
		if _, err := os.Stat("database.sqlite-wal"); err == nil {
			copyFile("database.sqlite-wal", filepath.Join(buildDir, "Storage", "database.sqlite-wal"))
		}
		fmt.Println("🗄️  Base de datos copiada a build/Storage/")
	}

	// Make binary executable on unix
	if targetOS != "windows" {
		os.Chmod(outPath, 0755)
	}

	stat, _ := os.Stat(outPath)
	sizeMB := float64(stat.Size()) / (1024 * 1024)

	fmt.Printf("\n✨ ¡COMPILACIÓN NATIVA EXITOSA!\n")
	fmt.Printf(" Archivo de Salida : %s\n", outPath)
	fmt.Printf(" Tamaño Binario   : %.2f MB\n", sizeMB)
	fmt.Printf(" Destino           : %s/%s\n", targetOS, targetArch)
	fmt.Printf(" Instrucciones    : Copia '%s' a cualquier PC con %s y ejecútalo directamente (¡no requiere Joss instalado!).\n\n", outPath, strings.ToUpper(targetOS))
}

func getWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "app"
	}
	return dir
}
