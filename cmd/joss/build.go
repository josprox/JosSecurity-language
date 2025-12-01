package main

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jossecurity/joss/pkg/crypto"
)

const runnerSource = `package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/jchv/go-webview2"
	"github.com/jossecurity/joss/pkg/crypto"
	"github.com/jossecurity/joss/pkg/server"
	"github.com/jossecurity/joss/pkg/vfs"

	_ "embed"
)

//go:embed assets.bin
var encryptedAssets []byte

// This variable is injected at build time via string replacement
const BuildKey = "BUILD_KEY_PLACEHOLDER"

func main() {
	// 1. Setup Logging to error.log
	setupLogging()

	// 2. Decrypt Assets
	if BuildKey == "BUILD_KEY_PLACEHOLDER" {
		log.Fatal("BuildKey not injected")
	}

	obfuscatedKey, err := hex.DecodeString(BuildKey)
	if err != nil {
		log.Fatalf("Error decoding BuildKey: %v", err)
	}

	key := crypto.DeobfuscateKey(obfuscatedKey) 

	decryptedData, err := crypto.DecryptAES(encryptedAssets, key)
	if err != nil {
		log.Fatalf("Error decrypting assets: %v", err)
	}

	// 3. Hydrate VFS
	var files map[string][]byte
	decDecoder := gob.NewDecoder(bytes.NewReader(decryptedData))
	if err := decDecoder.Decode(&files); err != nil {
		log.Fatalf("Error decoding assets: %v", err)
	}

	memFS := vfs.NewMemFS()
	memFS.Files = files

	// 4. Start Server with VFS
	go func() {
		server.Start(memFS)
	}()

	// 5. Wait for Server
	waitForServer("localhost:8000")

	// 6. Launch WebView
	w := webview2.New(true)
	if w == nil {
		log.Fatal("Failed to load WebView2.")
	}
	defer w.Destroy()

	w.SetTitle("JosSecurity App")
	w.SetSize(1024, 768, webview2.HintNone)
	w.Navigate("http://localhost:8000")
	w.Run()
}

func setupLogging() {
	exePath, err := os.Executable()
	if err != nil {
		exePath = "."
	}
	dir := filepath.Dir(exePath)
	logFile, err := os.OpenFile(filepath.Join(dir, "error.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
	}
}

func waitForServer(address string) {
	for i := 0; i < 30; i++ {
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}
`

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

	// 2. Prepare Build Directory
	buildDir := "build"
	tempDir := filepath.Join(buildDir, "temp")
	os.RemoveAll(buildDir)
	os.MkdirAll(filepath.Join(buildDir, "data"), 0755)
	os.MkdirAll(filepath.Join(buildDir, "Storage"), 0755)
	os.MkdirAll(tempDir, 0755)

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
			// If it's env.joss and we are building a program with a database,
			// we need to update the DB_PATH to point to the Storage folder.
			if f == "env.joss" {
				if _, err := os.Stat("database.sqlite"); err == nil {
					// Append the override
					// We use a newline to ensure it's on a new line
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

	// Write assets.bin to temp dir
	if err := ioutil.WriteFile(filepath.Join(tempDir, "assets.bin"), encryptedAssets, 0644); err != nil {
		fmt.Printf("Error escribiendo assets.bin: %v\n", err)
		return
	}

	// Prepare Runner Source with Injected Key
	obfuscatedKey := crypto.ObfuscateKey(buildKey)
	hexKey := fmt.Sprintf("%x", obfuscatedKey)

	// Inject Key via String Replacement
	finalRunnerSource := bytes.Replace([]byte(runnerSource), []byte("BUILD_KEY_PLACEHOLDER"), []byte(hexKey), 1)

	// Write main.go (runner) to temp dir
	if err := ioutil.WriteFile(filepath.Join(tempDir, "main.go"), finalRunnerSource, 0644); err != nil {
		fmt.Printf("Error escribiendo main.go: %v\n", err)
		return
	}

	// 4. Initialize Go Module in temp dir
	fmt.Println("Inicializando entorno de compilación...")
	runCmd(tempDir, "go", "mod", "init", "runner")

	// If we are in the JosSecurity repo (dev mode), replace the dependency
	if _, err := os.Stat("../../go.mod"); err == nil {
		absPath, _ := filepath.Abs("../..")
		fmt.Printf("Detectado entorno de desarrollo, usando reemplazo local: %s\n", absPath)
		runCmd(tempDir, "go", "mod", "edit", "-replace", "github.com/jossecurity/joss="+absPath)
	}

	runCmd(tempDir, "go", "mod", "tidy")

	// 5. Compile Runner
	fmt.Println("Compilando runner...")

	ldflags := "-s -w -H=windowsgui"
	if goos != "windows" {
		ldflags = "-s -w"
	}

	outPath := filepath.Join("..", "main"+ext) // ../main.exe -> build/main.exe

	cmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", outPath, ".")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "GOOS="+goos, "CGO_ENABLED=1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error compilando runner: %v\n%s\n", err, output)
		fmt.Println("Intentando sin CGO...")
		cmd = exec.Command("go", "build", "-ldflags", ldflags, "-o", outPath, ".")
		cmd.Dir = tempDir
		cmd.Env = append(os.Environ(), "GOOS="+goos, "CGO_ENABLED=0")
		output, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error fatal compilando runner: %v\n%s\n", err, output)
			return
		}
	}

	// Cleanup Temp
	os.RemoveAll(tempDir)

	// 6. Copy Database
	if _, err := os.Stat("database.sqlite"); err == nil {
		copyFile("database.sqlite", filepath.Join(buildDir, "Storage", "database.sqlite"))
		fmt.Println("Base de datos copiada a build/Storage/")
	}

	// 7. Create error.log
	ioutil.WriteFile(filepath.Join(buildDir, "error.log"), []byte(""), 0666)

	fmt.Println("Build PROGRAM completado exitosamente en carpeta 'build/'.")
	fmt.Println("Estructura:")
	fmt.Printf("  %s\n", filepath.Join(buildDir, "main"+ext))
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
