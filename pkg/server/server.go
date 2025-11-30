package server

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jossecurity/joss/pkg/core"
)

var (
	currentRuntime   *core.Runtime
	mutex            sync.RWMutex
	GlobalFileSystem http.FileSystem // Exposed for hotreload.go
)

// Start initializes and starts the Joss HTTP server with Hot Reload
// fs can be nil, in which case it defaults to http.Dir("public")
func Start(fileSystem http.FileSystem) {
	GlobalFileSystem = fileSystem
	core.SetFileSystem(fileSystem)

	// Initial Load
	reloadApp("")

	// Start File Watcher
	go watchChanges()

	port := "8000"
	if val, ok := currentRuntime.Env["PORT"]; ok && val != "" {
		port = val
	}

	// Static Files
	if GlobalFileSystem != nil {
		// VFS Mode (Root FS)
		fsHandler := http.FileServer(GlobalFileSystem)

		// /public/ -> maps to public/ in VFS (no strip prefix)
		http.Handle("/public/", fsHandler)

		// /assets/ -> maps to public/ in VFS
		http.Handle("/assets/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = strings.Replace(r.URL.Path, "/assets/", "/public/", 1)
			fsHandler.ServeHTTP(w, r)
		}))
	} else {
		// Disk Mode (Public FS)
		if fileSystem == nil {
			fileSystem = http.Dir("public")
		}
		fsHandler := http.FileServer(fileSystem)
		http.Handle("/public/", http.StripPrefix("/public/", fsHandler))
		http.Handle("/assets/", http.StripPrefix("/assets/", fsHandler))
	}

	// SSE Endpoint for Hot Reload
	http.HandleFunc("/__hot_reload", sseHandler)

	// WebSocket Endpoint
	InitWebSocket()
	core.BroadcastFunc = Broadcast
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(GlobalHub, w, r)
	})

	// Main Handler
	http.HandleFunc("/", MainHandler)

	// Start Server with Timeouts
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      nil, // Use DefaultServeMux
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("Iniciando servidor JosSecurity en http://localhost:%s\n", port)
	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("Error iniciando servidor: %v\n", err)
	}
}
