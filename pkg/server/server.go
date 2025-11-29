package server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/jossecurity/joss/pkg/core"
)

var (
	currentRuntime *core.Runtime
	mutex          sync.RWMutex
)

// Start initializes and starts the Joss HTTP server with Hot Reload
func Start() {
	// Initial Load
	reloadApp("")

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
	http.HandleFunc("/", MainHandler)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Error iniciando servidor: %v\n", err)
	}
}
