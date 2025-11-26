package server

import (
	"fmt"
	"net/http"

	"github.com/jossecurity/joss/pkg/core"
)

// Start initializes and starts the Joss HTTP server
func Start() {
	// Initialize Runtime
	rt := core.NewRuntime()
	rt.LoadEnv()

	port := rt.Env["PORT"]
	if port == "" {
		port = "8000"
	}

	fmt.Printf("Iniciando servidor JosSecurity en http://localhost:%s\n", port)
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "JosSecurity Server Running\n")
		fmt.Fprintf(w, "Environment: %s\n", rt.Env["APP_ENV"])
	})

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Error iniciando servidor: %v\n", err)
	}
}
