//go:build windows

package main

import (
	"log"
	"net"
	"time"

	"github.com/jchv/go-webview2"
	"github.com/jossecurity/joss/pkg/server"

	_ "embed"
)

//go:embed default_logo.png
var defaultLogo []byte

func startProgram() {
	// Start Server in Goroutine
	go func() {
		server.Start(nil)
	}()

	// Wait for server to be ready
	waitForServer("localhost:8000")

	// Create WebView2 instance
	w := webview2.New(true)
	if w == nil {
		log.Println("Failed to load WebView2. Is Edge installed?")
		return
	}
	defer w.Destroy()

	w.SetTitle("JosSecurity App")
	w.SetSize(1024, 768, webview2.HintNone)

	// Navigate to local server
	w.Navigate("http://localhost:8000")

	// Run the application
	w.Run()
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
