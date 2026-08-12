package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jossecurity/joss/pkg/bytecode"
	"github.com/jossecurity/joss/pkg/crypto"
	"github.com/jossecurity/joss/pkg/parser"
)

const (
	sqliteDbFile  = "database.sqlite"
	sqliteShmFile = "database.sqlite-shm"
	sqliteWalFile = "database.sqlite-wal"
)

var supportedTargets = map[string][]string{
	"windows": {"amd64", "arm64", "386"},
	"linux":   {"amd64", "arm64", "386"},
	"darwin":  {"amd64", "arm64"},
}

// buildNative builds a standalone native executable binary.
func buildNative(targetOS, targetArch string, enableGUI bool) {
	tOS, tArch, valid := validateBuildTarget(targetOS, targetArch)
	if !valid {
		return
	}

	modeStr := "Consola/Servidor Headless"
	if enableGUI {
		modeStr = "Interfaz Gráfica GUI (WebView2)"
	}

	fmt.Printf("\n=======================================================\n")
	fmt.Printf("🚀 COMPILADOR NATIVO DE JOSS (AOT & Standalone Mode)\n")
	fmt.Printf(" Target OS   : %s\n", strings.ToUpper(tOS))
	fmt.Printf(" Target Arch : %s\n", strings.ToUpper(tArch))
	fmt.Printf(" Modo GUI    : %s\n", modeStr)
	fmt.Printf(" Stripped    : -ldflags=\"-s -w\"\n")
	fmt.Printf(" CGO Enabled : 0 (Enlazado Estático Puro)\n")
	fmt.Printf("=======================================================\n\n")

	if _, err := exec.LookPath("go"); err != nil {
		fmt.Println("Error: No se encontró la herramienta 'go' instalada en el sistema.")
		return
	}

	buildDir := "build"
	os.RemoveAll(buildDir)
	if err := os.MkdirAll(filepath.Join(buildDir, "Storage"), 0755); err != nil {
		fmt.Printf("Error creando directorio build: %v\n", err)
		return
	}

	fmt.Println("📦 Empaquetando y precompilando assets (AOT Bytecode AST)...")
	encryptedAssets, buildKey, err := collectAndEncryptAssets(enableGUI)
	if err != nil {
		fmt.Printf("Error procesando assets del proyecto: %v\n", err)
		return
	}

	fmt.Println("🔨 Compilando ejecutable nativo con la toolchain de Go...")
	runnerBytes, err := compileRunnerBinary(tOS, tArch, enableGUI)
	if err != nil {
		fmt.Printf("Error compilando runner nativo: %v\n", err)
		return
	}

	outPath, err := assembleFinalExecutable(buildDir, tOS, runnerBytes, encryptedAssets, buildKey)
	if err != nil {
		fmt.Printf("Error ensamblando ejecutable final: %v\n", err)
		return
	}

	copyDatabaseFiles(buildDir)

	stat, _ := os.Stat(outPath)
	sizeMB := float64(stat.Size()) / (1024 * 1024)

	fmt.Printf("\n✨ ¡COMPILACIÓN NATIVA EXITOSA!\n")
	fmt.Printf(" Archivo de Salida : %s\n", outPath)
	fmt.Printf(" Tamaño Binario   : %.2f MB (Minificado & Comprimido)\n", sizeMB)
	fmt.Printf(" Destino           : %s/%s\n", tOS, tArch)
	fmt.Printf(" Modo              : %s\n", modeStr)
	fmt.Printf(" Instrucciones    : Copia '%s' a cualquier PC con %s y ejecútalo directamente.\n\n", outPath, strings.ToUpper(tOS))
}

func validateBuildTarget(targetOS, targetArch string) (string, string, bool) {
	if targetOS == "" {
		targetOS = runtime.GOOS
	}
	if targetArch == "" {
		targetArch = runtime.GOARCH
	}

	tOS := strings.ToLower(targetOS)
	tArch := strings.ToLower(targetArch)

	archs, osValid := supportedTargets[tOS]
	if !osValid {
		fmt.Printf("Error: Sistema operativo '%s' no soportado. Opciones: windows, linux, darwin.\n", targetOS)
		return tOS, tArch, false
	}

	for _, a := range archs {
		if a == tArch {
			return tOS, tArch, true
		}
	}

	fmt.Printf("Error: Arquitectura '%s' no soportada para %s. Opciones: %s\n", targetArch, targetOS, strings.Join(archs, ", "))
	return tOS, tArch, false
}

func collectAndEncryptAssets(enableGUI bool) ([]byte, []byte, error) {
	buildKey := make([]byte, 32)
	if _, err := rand.Read(buildKey); err != nil {
		return nil, nil, err
	}

	files := make(map[string][]byte)
	ignoredDirs := map[string]bool{
		".git": true, ".vscode": true, ".idea": true, "build": true, "vendor": true,
		"node_modules": true, ".gemini": true, ".codex": true, ".agents": true, ".github": true,
	}

	compiledCount := 0
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		return processSingleWalkFile(path, info, err, files, ignoredDirs, &compiledCount)
	})

	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("⚡ Pre-compilados %d archivos Joss a bytecode nativo.\n", compiledCount)
	encryptProjectEnvironment(files, enableGUI)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(files); err != nil {
		return nil, nil, err
	}

	encryptedAssets, err := crypto.EncryptAES(buf.Bytes(), buildKey)
	return encryptedAssets, buildKey, err
}

func processSingleWalkFile(path string, info os.FileInfo, err error, files map[string][]byte, ignoredDirs map[string]bool, compiledCount *int) error {
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
	if shouldSkipFileByName(name) || info.Size() > 5<<20 {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	relPath := filepath.ToSlash(path)
	if strings.HasSuffix(relPath, ".joss") {
		if bcData, ok := tryCompileJossBytecode(data); ok {
			data = bcData
			*compiledCount++
		}
	}

	files[relPath] = data
	return nil
}

func shouldSkipFileByName(name string) bool {
	return strings.HasSuffix(name, ".exe") || strings.HasSuffix(name, ".log") ||
		strings.HasSuffix(name, ".enc") || strings.HasSuffix(name, ".dll") ||
		strings.HasSuffix(name, ".so") || strings.HasSuffix(name, ".dylib") ||
		name == "runner" || name == "joss"
}

func tryCompileJossBytecode(data []byte) ([]byte, bool) {
	l := parser.NewLexer(string(data))
	p := parser.NewParser(l)
	prog := p.ParseProgram()
	if len(p.Errors()) == 0 && prog != nil {
		if bc, bcErr := bytecode.Encode(prog); bcErr == nil {
			return bc, true
		}
	}
	return nil, false
}

func encryptProjectEnvironment(files map[string][]byte, enableGUI bool) {
	envPath := GetEnvFile()
	data, _ := os.ReadFile(envPath)

	if enableGUI {
		data = append(data, []byte("\nJOSS_GUI=\"true\"")...)
	} else {
		data = append(data, []byte("\nJOSS_GUI=\"false\"")...)
	}

	if _, err := os.Stat(sqliteDbFile); err == nil {
		override := "\nDB_PATH=\"Storage/" + sqliteDbFile + "\""
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

func compileRunnerBinary(targetOS, targetArch string, enableGUI bool) ([]byte, error) {
	tempRunnerDir, err := os.MkdirTemp("", "joss-build-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempRunnerDir)

	tempRunnerBin := filepath.Join(tempRunnerDir, "runner")
	if targetOS == "windows" {
		tempRunnerBin += ".exe"
	}

	runnerPkg := "github.com/jossecurity/joss/cmd/runner"
	if _, err := os.Stat("cmd/runner"); err == nil {
		runnerPkg = "./cmd/runner"
	}

	ldflags := "-s -w"
	if targetOS == "windows" && enableGUI {
		ldflags += " -H=windowsgui"
	}
	cmd := exec.Command("go", "build", "-ldflags="+ldflags, "-o", tempRunnerBin, runnerPkg)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+targetOS, "GOARCH="+targetArch)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, string(out))
	}

	if upxPath, err := exec.LookPath("upx"); err == nil {
		upxCmd := exec.Command(upxPath, "--best", "--lzma", tempRunnerBin)
		_ = upxCmd.Run()
	}

	return os.ReadFile(tempRunnerBin)
}

func compressFinalExecutableWithUPX(outPath string) {
	if upxPath, err := exec.LookPath("upx"); err == nil {
		upxCmd := exec.Command(upxPath, "--best", "--lzma", outPath)
		_ = upxCmd.Run()
	}
}

func assembleFinalExecutable(buildDir, targetOS string, runnerBytes, encryptedAssets, buildKey []byte) (string, error) {
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
		return "", err
	}
	defer outFile.Close()

	if _, err := outFile.Write(runnerBytes); err != nil {
		return "", err
	}
	if _, err := outFile.Write(encryptedAssets); err != nil {
		return "", err
	}
	if _, err := outFile.Write(buildKey); err != nil {
		return "", err
	}

	lenBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(lenBuf, uint64(len(encryptedAssets)))
	if _, err := outFile.Write(lenBuf); err != nil {
		return "", err
	}

	magic := []byte("JOSS_RUNNER_DATA")
	if _, err := outFile.Write(magic); err != nil {
		return "", err
	}

	if targetOS != "windows" {
		os.Chmod(outPath, 0755)
	}

	compressFinalExecutableWithUPX(outPath)

	return outPath, nil
}

func copyDatabaseFiles(buildDir string) {
	if _, err := os.Stat(sqliteDbFile); err == nil {
		copyFile(sqliteDbFile, filepath.Join(buildDir, "Storage", sqliteDbFile))
		if _, err := os.Stat(sqliteShmFile); err == nil {
			copyFile(sqliteShmFile, filepath.Join(buildDir, "Storage", sqliteShmFile))
		}
		if _, err := os.Stat(sqliteWalFile); err == nil {
			copyFile(sqliteWalFile, filepath.Join(buildDir, "Storage", sqliteWalFile))
		}
		fmt.Println("🗄️  Base de datos copiada a build/Storage/")
	}
}

func getWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "app"
	}
	return dir
}
