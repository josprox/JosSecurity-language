package main

import (
	"bytes"
	"encoding/gob"
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

var encryptedAssets []byte

// These variables are injected at build time via -ldflags
var (
	BuildSalt string // Hex encoded salt
	BuildKey  string // Hex encoded obfuscated key
)

func main() {
	// 1. Setup Logging to error.log
	setupLogging()

	// 2. Decrypt Assets
	// In a real scenario, we would deobfuscate BuildKey and use BuildSalt
	// For this prototype, we assume the key is passed or hardcoded for now if injection fails
	// But the goal is to use the injected values.

	// TODO: Implement proper key deobfuscation and salt usage from ldflags
	// For now, we will assume a fixed key for the prototype to ensure it works first
	// key := crypto.DeobfuscateKey([]byte(BuildKey))

	// Placeholder key for development (must match build.go)
	key := []byte("12345678901234567890123456789012")

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
