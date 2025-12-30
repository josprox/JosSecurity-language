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
	"github.com/jossecurity/joss/pkg/core"
	"github.com/jossecurity/joss/pkg/crypto"
	"github.com/jossecurity/joss/pkg/parser"
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

	// Fix: Change CWD to executable directory to ensure relative paths (like Storage/database.sqlite) work
	exeDir := filepath.Dir(exePath)
	if err := os.Chdir(exeDir); err != nil {
		log.Printf("Warning: Could not change CWD to %s: %v", exeDir, err)
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

	// 4. Handle Arguments (Support for self-spawned "server start")
	if len(os.Args) >= 3 && os.Args[1] == "server" && os.Args[2] == "start" {
		server.Start(memFS)
		return
	}

	// 5. Normal Startup: Execute main.joss and Open WebView
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in runtime: %v", r)
			}
		}()

		// Set Global FileSystem for Server::start() native calls
		server.GlobalFileSystem = memFS
		core.SetFileSystem(memFS)

		r := core.NewRuntime()
		r.LoadEnv(memFS)

		// Execute main.joss
		content, err := memFS.Open("main.joss")
		if err == nil {
			stat, _ := content.Stat()
			data := make([]byte, stat.Size())
			content.Read(data)
			content.Close()

			l := parser.NewLexer(string(data))
			p := parser.NewParser(l)
			program := p.ParseProgram()
			if len(p.Errors()) == 0 {
				r.Execute(program)
			} else {
				log.Printf("Parser Errors in main.joss: %v", p.Errors())
			}
		} else {
			// Fallback: Start server directly
			server.Start(memFS)
		}
	}()

	// Determine port from env (loaded from VFS or defaults)
	port := "8000"
	if envData, ok := files["env.joss"]; ok {
		// Simple parse
		lines := bytes.Split(envData, []byte("\n"))
		for _, line := range lines {
			s := string(bytes.TrimSpace(line))
			if (len(s) > 5 && s[:5] == "PORT=") || (len(s) > 10 && s[:10] == "JOSS_PORT=") {
				parts := bytes.SplitN(line, []byte("="), 2)
				if len(parts) == 2 {
					val := bytes.TrimSpace(parts[1])
					val = bytes.Trim(val, "\"")
					val = bytes.Trim(val, "'")
					port = string(val)
				}
			}
		}
	} else if envEnc, ok := files["env.enc"]; ok {
		if len(envEnc) > 16 {
			salt := envEnc[:16]
			ciphertext := envEnc[16:]
			masterSecret := []byte("JOSSECURITY_MASTER_SECRET_2025")
			key := crypto.DeriveKey(masterSecret, salt)
			decrypted, err := crypto.DecryptAES(ciphertext, key)
			if err == nil {
				lines := bytes.Split(decrypted, []byte("\n"))
				for _, line := range lines {
					s := string(bytes.TrimSpace(line))
					if (len(s) > 5 && s[:5] == "PORT=") || (len(s) > 10 && s[:10] == "JOSS_PORT=") {
						parts := bytes.SplitN(line, []byte("="), 2)
						if len(parts) == 2 {
							val := bytes.TrimSpace(parts[1])
							val = bytes.Trim(val, "\"")
							val = bytes.Trim(val, "'")
							port = string(val)
						}
					}
				}
			}
		}
	}

	// Wait for resolved port, or fallback to default
	finalPort := waitForPortOrPort(port, "8000")

	w := webview2.New(true)
	if w == nil {
		log.Println("Failed to load WebView2.")
		return
	}
	defer w.Destroy()

	w.SetTitle("JosSecurity App")
	w.SetSize(1024, 768, webview2.HintNone)
	w.Navigate("http://localhost:" + finalPort)
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
		// Redirect Stdout and Stderr to capture fmt.Println from runtime
		// Note: This works for fmt package but not necessarily for low-level writes,
		// but for our runtime debugging it's sufficient.
		// Actually, assigning to os.Stdout/Stderr is not thread-safe and might not work as expected
		// if runtime uses the original file descriptors.
		// A better way is to replace the file descriptors, but that's OS specific.
		// For simplicity in Go, we can just set os.Stdout = logFile.
		os.Stdout = logFile
		os.Stderr = logFile
	}
}

func waitForPortOrPort(p1, p2 string) string {
	target := ""
	for i := 0; i < 60; i++ { // 12 seconds
		if checkPort(p1) {
			target = p1
			break
		}
		if checkPort(p2) {
			target = p2
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if target == "" {
		return p1 // Default to first
	}
	return target
}

func checkPort(port string) bool {
	conn, err := net.DialTimeout("tcp", "localhost:"+port, 100*time.Millisecond)
	if err == nil {
		conn.Close()
		return true
	}
	return false
}
