package terminal

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gorilla/websocket"
	"github.com/creack/pty"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// TerminalSession represents a WebSocket terminal session
type TerminalSession struct {
	conn   *websocket.Conn
	cmd    *exec.Cmd
	pty    *os.File
}

// handleWebSocket handles WebSocket connections for the terminal
func WebSocket(w http.ResponseWriter, r *http.Request) {
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

// startShell starts a new shell process with PTY
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
	
	log.Printf("PTY created, waiting for client to set size")
	
	return nil
}

// handle manages the WebSocket communication
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
					log.Printf("Resizing PTY to %dx%d", resizeMsg.Cols, resizeMsg.Rows)
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

// cleanup cleans up the terminal session
func (ts *TerminalSession) cleanup() {
	if ts.pty != nil {
		ts.pty.Close()
	}
	if ts.cmd != nil && ts.cmd.Process != nil {
		ts.cmd.Process.Kill()
	}
} 