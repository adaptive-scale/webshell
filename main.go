package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/creack/pty"
	"html/template"
)

// CommandResponse represents the structure of command execution responses
type CommandResponse struct {
	Success   bool   `json:"success"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
	ExitCode  int    `json:"exit_code"`
	Duration  string `json:"duration"`
	Timestamp string `json:"timestamp"`
	Command   string `json:"command"`
}

// AllowedCommands defines which commands are allowed to be executed
var AllowedCommands = map[string]bool{
	"ls":      true,
	"pwd":     true,
	"whoami":  true,
	"date":    true,
	"uptime":  true,
	"ps":      true,
	"df":      true,
	"free":    true,
	"top":     true,
	"cat":     true,
	"head":    true,
	"tail":    true,
	"grep":    true,
	"find":    true,
	"echo":    true,
	"uname":   true,
	"hostname": true,
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Terminal session
type TerminalSession struct {
	conn   *websocket.Conn
	cmd    *exec.Cmd
	pty    *os.File
}

func main() {
	port := getEnv("PORT", "8080")
	
	// Set up routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/execute", handleExecuteCommand)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/terminal", handleTerminalPage)
	http.HandleFunc("/ws", handleWebSocket)
	
	log.Printf("Starting HTTP server on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("  GET  / - Home page with usage information")
	log.Printf("  POST /execute - Execute a command (raw body)")
	log.Printf("  GET  /health - Health check")
	log.Printf("  GET  /terminal - Web SSH terminal")
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.ParseFiles("webhttpexec/templates/home.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	// Convert AllowedCommands map to slice for template iteration
	var allowedCommands []string
	for cmd := range AllowedCommands {
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

func handleExecuteCommand(w http.ResponseWriter, r *http.Request) {
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
	if !AllowedCommands[command] {
		http.Error(w, "Command not allowed", http.StatusForbidden)
		return
	}
	
	// Check if JSON response is requested
	acceptHeader := r.Header.Get("Accept")
	wantJSON := acceptHeader == "application/json"
	
	// Execute command
	response := executeCommand(command, args)
	
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

func handleHealth(w http.ResponseWriter, r *http.Request) {
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

func executeCommand(command string, args []string) CommandResponse {
	start := time.Now()
	
	// Create command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	
	// Execute command
	output, err := cmd.CombinedOutput()
	
	duration := time.Since(start)
	
	response := CommandResponse{
		Success:   err == nil,
		Output:    strings.TrimSpace(string(output)),
		ExitCode:  cmd.ProcessState.ExitCode(),
		Duration:  duration.String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Command:   command + " " + strings.Join(args, " "),
	}
	
	if err != nil {
		response.Error = err.Error()
	}
	
	return response
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func handleTerminalPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.ParseFiles("webhttpexec/templates/terminal.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Hostname string
	}{
		Hostname: getHostname(),
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()
	
	// Create a new terminal session
	session := &TerminalSession{
		conn: conn,
	}
	
	// Start the shell process
	if err := session.startShell(); err != nil {
		log.Printf("Failed to start shell: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error: Failed to start shell\r\n"))
		return
	}
	defer session.cleanup()
	
	// Handle terminal session
	session.handle()
}

func (ts *TerminalSession) startShell() error {
	// Start bash shell with proper environment
	ts.cmd = exec.Command("bash")
	
	// Set environment variables for proper terminal support
	ts.cmd.Env = append(os.Environ(),
		"TERM=xterm",
		"TERMINFO=/usr/share/terminfo",
	)
	
	// Create PTY
	ptyFile, err := pty.Start(ts.cmd)
	if err != nil {
		return err
	}
	ts.pty = ptyFile
	
	// Don't set initial size - let client set it based on actual screen size
	log.Printf("PTY created, waiting for client to set size")
	
	return nil
}

func (ts *TerminalSession) handle() {
	// Channel to signal when the command is done
	done := make(chan bool)
	
	// Read from PTY and send to WebSocket
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := ts.pty.Read(buffer)
			if err != nil {
				if err != io.EOF {
					log.Printf("Error reading from PTY: %v", err)
				}
				break
			}
			if n > 0 {
				err = ts.conn.WriteMessage(websocket.TextMessage, buffer[:n])
				if err != nil {
					log.Printf("Error writing to WebSocket: %v", err)
					break
				}
			}
		}
		done <- true
	}()
	
	// Read from WebSocket and write to PTY
	go func() {
		for {
			_, message, err := ts.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}
			
			// Check if it's a resize message
			if len(message) > 0 && message[0] == '{' {
				var resizeMsg struct {
					Type string `json:"type"`
					Cols int    `json:"cols"`
					Rows int    `json:"rows"`
				}
				if err := json.Unmarshal(message, &resizeMsg); err == nil && resizeMsg.Type == "resize" {
					// Resize the PTY
					log.Printf("Resizing PTY to %dx%d (screen-based size)", resizeMsg.Cols, resizeMsg.Rows)
					if err := pty.Setsize(ts.pty, &pty.Winsize{
						Rows: uint16(resizeMsg.Rows),
						Cols: uint16(resizeMsg.Cols),
					}); err != nil {
						log.Printf("Error resizing PTY: %v", err)
					} else {
						log.Printf("Successfully resized PTY to %dx%d", resizeMsg.Cols, resizeMsg.Rows)
						
						// Update environment variables in the shell
						colsStr := fmt.Sprintf("%d", resizeMsg.Cols)
						rowsStr := fmt.Sprintf("%d", resizeMsg.Rows)
						
						// Send commands to update COLUMNS and LINES
						updateCmd := fmt.Sprintf("export COLUMNS=%s; export LINES=%s; stty cols %s rows %s\n", 
							colsStr, rowsStr, colsStr, rowsStr)
						ts.pty.Write([]byte(updateCmd))
					}
					continue
				}
			}
			
			// Write to PTY
			_, err = ts.pty.Write(message)
			if err != nil {
				log.Printf("Error writing to PTY: %v", err)
				break
			}
		}
		done <- true
	}()
	
	// Wait for the command to finish
	ts.cmd.Wait()
	<-done
	<-done
}

func (ts *TerminalSession) cleanup() {
	if ts.pty != nil {
		ts.pty.Close()
	}
	if ts.cmd != nil && ts.cmd.Process != nil {
		ts.cmd.Process.Kill()
	}
}
