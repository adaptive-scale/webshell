package handler

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"github.com/adaptive-scale/webshell/internal/commands"
	"github.com/adaptive-scale/webshell/internal/config"

)

// handleHome serves the home page with usage information
func Home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	// Convert AllowedCommands map to slice for template iteration
	var allowedCommands []string
	for cmd := range commands.AllowedCommands {
		allowedCommands = append(allowedCommands, cmd)
	}

	data := struct {
		AllowedCommands []string
	}{
		AllowedCommands: allowedCommands,
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}

// handleExecuteCommand executes commands via HTTP POST
func ExecuteCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Read raw body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	// Parse command from raw body
	commandLine := strings.TrimSpace(string(body))
	if commandLine == "" {
		http.Error(w, "Command is required", http.StatusBadRequest)
		return
	}
	
	// Split command into command and arguments
	parts := strings.Fields(commandLine)
	if len(parts) == 0 {
		http.Error(w, "Invalid command", http.StatusBadRequest)
		return
	}
	
	command := parts[0]
	args := parts[1:]
	
	// Validate command
	if !commands.AllowedCommands[command] {
		http.Error(w, "Command not allowed", http.StatusForbidden)
		return
	}
	
	// Check if JSON response is requested
	acceptHeader := r.Header.Get("Accept")
	wantJSON := acceptHeader == "application/json"
	
	// Execute command
	response := commands.ExecuteCommand(command, args)
	
	// Return response based on Accept header
	if wantJSON {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	} else {
		// Return raw output
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		if response.Success {
			w.Write([]byte(response.Output))
		} else {
			w.Write([]byte("Error: " + response.Error))
		}
	}
}

// handleHealth serves the health check endpoint
func Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    "running",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleTerminalPage serves the web terminal page
func TerminalPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.ParseFiles("templates/terminal.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Hostname string
	}{
		Hostname: config.GetHostname(),
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
} 