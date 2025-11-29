package server

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jossecurity/joss/pkg/core"
	"github.com/jossecurity/joss/pkg/parser"
)

var (
	sseClients   = make(map[chan string]bool)
	sseClientsMu sync.Mutex
)

func sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientChan := make(chan string)
	sseClientsMu.Lock()
	sseClients[clientChan] = true
	sseClientsMu.Unlock()

	defer func() {
		sseClientsMu.Lock()
		delete(sseClients, clientChan)
		sseClientsMu.Unlock()
		close(clientChan)
	}()

	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case msg := <-clientChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			w.(http.Flusher).Flush()
		}
	}
}

func notifyClients() {
	sseClientsMu.Lock()
	defer sseClientsMu.Unlock()
	for client := range sseClients {
		select {
		case client <- "reload":
		default:
		}
	}
}

func watchChanges() {
	// Store file hashes: path -> hash
	fileHashes := make(map[string]string)

	// Helper to calculate hash
	getHash := func(path string) string {
		content, err := os.ReadFile(path)
		if err != nil {
			return ""
		}
		// Simple MD5 hash
		hash := md5.Sum(content)
		return hex.EncodeToString(hash[:])
	}

	// Initial scan
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			ext := filepath.Ext(path)
			if ext == ".joss" || ext == ".html" || ext == ".css" || ext == ".js" || ext == ".scss" {
				fileHashes[path] = getHash(path)
			}
		}
		return nil
	})

	for {
		time.Sleep(500 * time.Millisecond)

		var changedPath string
		err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				if info.Name() == ".git" || info.Name() == "vendor" || info.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}

			ext := filepath.Ext(path)
			if ext == ".joss" || ext == ".html" || ext == ".css" || ext == ".js" || ext == ".scss" {
				currentHash := getHash(path)
				if lastHash, ok := fileHashes[path]; ok {
					if currentHash != lastHash {
						fileHashes[path] = currentHash
						changedPath = path
						fmt.Printf("[HotReload] Cambio detectado en: %s\n", path)
					}
				} else {
					// New file
					fileHashes[path] = currentHash
					changedPath = path
					fmt.Printf("[HotReload] Nuevo archivo detectado: %s\n", path)
				}
			}
			return nil
		})

		if err == nil && changedPath != "" {
			// Debounce
			time.Sleep(100 * time.Millisecond)
			reloadApp(changedPath)
		}
	}
}

func reloadApp(changedFile string) {
	mutex.Lock()
	defer mutex.Unlock()

	if changedFile == "" {
		fmt.Println("Recargando aplicación completa...")
	} else {
		fmt.Printf("Recargando parcial: %s\n", changedFile)
	}

	// 1. Styles
	if changedFile == "" || strings.HasSuffix(changedFile, ".scss") || strings.HasSuffix(changedFile, ".css") {
		compileStyles()
		if strings.HasSuffix(changedFile, ".scss") || strings.HasSuffix(changedFile, ".css") {
			notifyClients()
			return
		}
	}

	// 2. Views (HTML)
	if strings.HasSuffix(changedFile, ".html") {
		// Views are read from disk, so just notify
		notifyClients()
		return
	}

	// 3. Runtime Logic
	if currentRuntime == nil {
		currentRuntime = core.NewRuntime()
		currentRuntime.LoadEnv()

		// Init Redis if configured
		if currentRuntime.Env["SESSION_DRIVER"] == "redis" {
			host := "localhost:6379"
			if val, ok := currentRuntime.Env["REDIS_HOST"]; ok {
				host = val
			}
			pass := ""
			if val, ok := currentRuntime.Env["REDIS_PASSWORD"]; ok {
				pass = val
			}
			core.InitRedis(host, pass, 0)
			fmt.Println("[Security] Redis conectado para sesiones")
		}
	}

	// Helper to load a single file
	loadFile := func(path string) {
		fmt.Printf("[DEBUG] Loading file: %s\n", path)
		content, err := os.ReadFile(path)
		if err == nil {
			l := parser.NewLexer(string(content))
			p := parser.NewParser(l)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				fmt.Printf("[DEBUG] Parser errors in %s:\n", path)
				for _, msg := range p.Errors() {
					fmt.Printf("\t%s\n", msg)
				}
			}
			currentRuntime.Execute(program)
		} else {
			fmt.Printf("[DEBUG] Error reading %s: %v\n", path, err)
		}
	}

	if changedFile != "" && strings.HasSuffix(changedFile, ".joss") {
		if strings.HasSuffix(changedFile, "routes.joss") {
			// Reload Routes
			fmt.Println("[DEBUG] Reloading routes...")
			// Clear existing routes? Router overwrites, so it's okay.
			// Ideally we should clear, but Runtime doesn't expose a ClearRoutes method.
			// We can manually clear if we want, but overwriting is fine for now.
			if currentRuntime.Routes != nil {
				// Optional: clear routes to remove deleted ones
				currentRuntime.Routes = make(map[string]map[string]interface{})
			}
			loadFile(changedFile)
		} else {
			// Reload Controller/Model
			loadFile(changedFile)
		}
		notifyClients()
		return
	}

	// Full Reload (Initial or unknown change)
	if changedFile == "" {
		// Reset Runtime
		if currentRuntime != nil {
			currentRuntime.Free()
		}
		currentRuntime = core.NewRuntime()
		currentRuntime.LoadEnv()

		// Init Redis if configured
		if currentRuntime.Env["SESSION_DRIVER"] == "redis" {
			host := "localhost:6379"
			if val, ok := currentRuntime.Env["REDIS_HOST"]; ok {
				host = val
			}
			pass := ""
			if val, ok := currentRuntime.Env["REDIS_PASSWORD"]; ok {
				pass = val
			}
			core.InitRedis(host, pass, 0)
			fmt.Println("[Security] Redis conectado para sesiones")
		}

		// Load App Files
		err := filepath.Walk("app", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".joss") {
				loadFile(path)
			}
			return nil
		})
		if err != nil {
			fmt.Printf("[DEBUG] Error walking app directory: %v\n", err)
		}

		// Load Routes
		routesPath := "routes.joss"
		if _, err := os.Stat(routesPath); err == nil {
			loadFile(routesPath)
		} else {
			fmt.Println("[DEBUG] routes.joss not found")
		}
		notifyClients()
	}
}
