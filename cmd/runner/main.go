package main

import (
	"bytes"
	"encoding/binary"
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
)

const MagicMarker = "JOSS_RUNNER_DATA"

func main() {
	setupLogging()

	// 1. Read Assets from Tail
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
	}

	f, err := os.Open(exePath)
	if err != nil {
		log.Fatalf("Error opening executable: %v", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		log.Fatalf("Error getting file stat: %v", err)
	}
	fileSize := stat.Size()

	// Read Magic Marker (16 bytes)
	if fileSize < 16 {
		log.Fatal("Invalid executable size")
	}
	marker := make([]byte, 16)
	f.ReadAt(marker, fileSize-16)

	if string(marker) != MagicMarker {
		log.Fatal("Error: Corrupted or invalid runner binary (Magic Marker not found)")
	}

	// Read Assets Length (8 bytes)
	// Layout: [Data] [Key 32] [Len 8] [Magic 16]
	lenBuf := make([]byte, 8)
	f.ReadAt(lenBuf, fileSize-16-8)
	var assetsLen int64
	binary.Read(bytes.NewReader(lenBuf), binary.LittleEndian, &assetsLen)

	// Read Key (32 bytes)
	key := make([]byte, 32)
	f.ReadAt(key, fileSize-16-8-32)

	// Read Encrypted Assets
	encryptedAssets := make([]byte, assetsLen)
	f.ReadAt(encryptedAssets, fileSize-16-8-32-assetsLen)

	// 2. Decrypt Assets
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
