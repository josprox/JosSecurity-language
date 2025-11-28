package server

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jossecurity/joss/pkg/core"
	"github.com/jossecurity/joss/pkg/parser"
)

var (
	currentRuntime *core.Runtime
	mutex          sync.RWMutex
	clients        = make(map[chan string]bool)
	clientsMu      sync.Mutex
	sessionStore   = make(map[string]map[string]interface{})
	sessionMu      sync.Mutex
)

// Start initializes and starts the Joss HTTP server with Hot Reload
func Start() {
	// Initial Load
	reloadApp()

	// Start File Watcher
	go watchChanges()

	port := "8000"
	if val, ok := currentRuntime.Env["PORT"]; ok && val != "" {
		port = val
	}

	fmt.Printf("Iniciando servidor JosSecurity en http://localhost:%s\n", port)

	// Static Files
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// SSE Endpoint for Hot Reload
	http.HandleFunc("/__hot_reload", sseHandler)

	// WebSocket Endpoint
	InitWebSocket()
	core.BroadcastFunc = Broadcast
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(GlobalHub, w, r)
	})

	// Main Handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mutex.RLock()
		rt := currentRuntime
		mutex.RUnlock()

		// 1. Parse Request Data
		r.ParseMultipartForm(10 << 20) // 10MB
		reqData := make(map[string]interface{})

		// Query Params
		for k, v := range r.URL.Query() {
			if len(v) > 0 {
				reqData[k] = v[0]
			}
		}
		// Form Data
		for k, v := range r.PostForm {
			if len(v) > 0 {
				reqData[k] = v[0]
			}
		}
		// Method
		reqData["_method"] = r.Method
		reqData["_referer"] = r.Referer()

		// 2. Session Management
		sessionID := ""
		cookie, err := r.Cookie("joss_session")
		if err != nil {
			sessionID = generateSessionID()
			http.SetCookie(w, &http.Cookie{Name: "joss_session", Value: sessionID, Path: "/", HttpOnly: true})
		} else {
			sessionID = cookie.Value
		}

		sessionMu.Lock()
		if _, ok := sessionStore[sessionID]; !ok {
			sessionStore[sessionID] = make(map[string]interface{})
		}
		sessData := sessionStore[sessionID]
		sessionMu.Unlock()

		// Dispatch
		fmt.Printf("[DEBUG] Dispatching %s %s\n", r.Method, r.URL.Path)
		result, err := rt.Dispatch(r.Method, r.URL.Path, reqData, sessData)
		if err == nil {
			fmt.Printf("[DEBUG] Dispatch success. Result type: %T\n", result)
			// Handle Redirect (Map)
			if resMap, ok := result.(map[string]interface{}); ok {
				if val, ok := resMap["_type"]; ok && val == "REDIRECT" {
					http.Redirect(w, r, resMap["url"].(string), http.StatusFound)
					return
				}
			}
			// Handle Redirect (Instance - e.g. RedirectResponse)
			if resInst, ok := result.(*core.Instance); ok {
				if val, ok := resInst.Fields["_type"]; ok && val == "REDIRECT" {
					// Handle Flash Data
					if flash, ok := resInst.Fields["flash"].(map[string]interface{}); ok {
						sessionMu.Lock()
						if _, ok := sessionStore[sessionID]; !ok {
							sessionStore[sessionID] = make(map[string]interface{})
						}
						// Merge flash into session
						for k, v := range flash {
							sessionStore[sessionID][k] = v
						}
						sessionMu.Unlock()
					}

					http.Redirect(w, r, resInst.Fields["url"].(string), http.StatusFound)
					return
				}
			}

			// If result is string, write it
			if str, ok := result.(string); ok {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(str))
				// ... hot reload script ...
				fmt.Fprintf(w, `<script>
					const evtSource = new EventSource("/__hot_reload");
					evtSource.onmessage = function(event) {
						if (event.data === "reload") {
							console.log("Reloading...");
							location.reload();
						}
					};
				</script>`)
				return
			} else {
				fmt.Println("[DEBUG] Result is not a string or redirect map")
			}
		} else {
			fmt.Printf("[DEBUG] Dispatch error: %v\n", err)
		}

		// Fallback / 404
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, "<h1>JosSecurity Server Running</h1>")
			fmt.Fprintf(w, "<p>Environment: %s</p>", rt.Env["APP_ENV"])
			fmt.Fprintf(w, `<link rel="stylesheet" href="/public/css/app.css">`)

			// Hot Reload Script
			fmt.Fprintf(w, `<script>
				const evtSource = new EventSource("/__hot_reload");
				evtSource.onmessage = function(event) {
					if (event.data === "reload") {
						console.log("Reloading...");
						location.reload();
					}
				};
			</script>`)
			return
		}

		http.NotFound(w, r)
	})

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Error iniciando servidor: %v\n", err)
	}
}

func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func reloadApp() {
	mutex.Lock()
	defer mutex.Unlock()

	fmt.Println("Recargando aplicación...")
	compileStyles()

	// Get new runtime from pool
	if currentRuntime != nil {
		currentRuntime.Free()
	}
	currentRuntime = core.NewRuntime()
	currentRuntime.LoadEnv()

	// Load App Files (Controllers, Models, etc.)
	err := filepath.Walk("app", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".joss") {
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
				// Execute to register classes, ignore Main not found
				currentRuntime.Execute(program)
			} else {
				fmt.Printf("[DEBUG] Error reading %s: %v\n", path, err)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("[DEBUG] Error walking app directory: %v\n", err)
	}

	// Load Routes
	routesPath := "routes.joss"
	if _, err := os.Stat(routesPath); err == nil {
		fmt.Println("[DEBUG] Loading routes from routes.joss")
		content, err := os.ReadFile(routesPath)
		if err == nil {
			l := parser.NewLexer(string(content))
			p := parser.NewParser(l)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				fmt.Printf("[DEBUG] Parser errors in routes.joss:\n")
				for _, msg := range p.Errors() {
					fmt.Printf("\t%s\n", msg)
				}
			}
			currentRuntime.Execute(program)
		} else {
			fmt.Printf("[DEBUG] Error reading routes.joss: %v\n", err)
		}
	} else {
		fmt.Println("[DEBUG] routes.joss not found")
	}

	notifyClients()
}

func compileStyles() {
	fmt.Println("Compilando estilos...")
	// Find .scss files
	files, _ := filepath.Glob("assets/css/*.scss")
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		content := string(data)

		// 1. Extract Variables
		vars := make(map[string]string)
		reVar := regexp.MustCompile(`(\$\w+):\s*(.+?);`)
		content = reVar.ReplaceAllStringFunc(content, func(match string) string {
			parts := reVar.FindStringSubmatch(match)
			vars[parts[1]] = parts[2]
			return "" // Remove variable definition from output
		})

		// 2. Replace Variables
		for k, v := range vars {
			// Replace $var with value
			content = strings.ReplaceAll(content, k, v)
		}

		// 3. Simple Nesting (Very basic: selector { & inner { } })
		// Too complex for regex. Let's stick to variables for PoC.

		// Save to public/css/
		outFile := filepath.Join("public", "css", filepath.Base(file))
		outFile = strings.Replace(outFile, ".scss", ".css", 1)
		os.MkdirAll(filepath.Dir(outFile), 0755)
		os.WriteFile(outFile, []byte(content), 0644)
	}

	// Copy .css files
	cssFiles, _ := filepath.Glob("assets/css/*.css")
	for _, file := range cssFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		outFile := filepath.Join("public", "css", filepath.Base(file))
		os.MkdirAll(filepath.Dir(outFile), 0755)
		os.WriteFile(outFile, data, 0644)
	}
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientChan := make(chan string)
	clientsMu.Lock()
	clients[clientChan] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, clientChan)
		clientsMu.Unlock()
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
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client := range clients {
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

		changed := false
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
						changed = true
						fmt.Printf("[HotReload] Cambio detectado en: %s\n", path)
					}
				} else {
					// New file
					fileHashes[path] = currentHash
					changed = true
					fmt.Printf("[HotReload] Nuevo archivo detectado: %s\n", path)
				}
			}
			return nil
		})

		if err == nil && changed {
			// Debounce: Wait a bit to see if more changes come (e.g. "Save All")
			time.Sleep(100 * time.Millisecond)
			reloadApp()
		}
	}
}
