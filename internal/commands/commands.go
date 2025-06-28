package commands

import (
	"context"
	"os/exec"
	"strings"
	"time"
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

// executeCommand executes a command with timeout and returns the result
func ExecuteCommand(command string, args []string) CommandResponse {
	start := time.Now()
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Create command
	cmd := exec.CommandContext(ctx, command, args...)
	
	// Execute command
	output, err := cmd.CombinedOutput()
	
	// Calculate duration
	duration := time.Since(start)
	
	// Prepare response
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