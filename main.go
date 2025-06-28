package main

import (
	"log"
	"net/http"
	"github.com/adaptive-scale/webshell/internal/handler"
	"github.com/adaptive-scale/webshell/internal/terminal"
	"github.com/adaptive-scale/webshell/internal/config"

)

func main() {
	port := config.GetEnv("PORT", "8080")
	setupRoutes()
	
	log.Printf("WebShell server starting on port %s", port)
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed:", err)
	}
}

func setupRoutes() {
	http.HandleFunc("/", handler.Home)
	http.HandleFunc("/execute", handler.ExecuteCommand)
	http.HandleFunc("/health", handler.Health)
	http.HandleFunc("/terminal", handler.TerminalPage)
	http.HandleFunc("/ws", terminal.WebSocket)
}
