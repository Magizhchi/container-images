package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MATLABServer implements the MCP server for MATLAB code execution
type MATLABServer struct {
	matlabPath string
}

// NewMATLABServer creates a new MATLAB Batch MCP server instance
func NewMATLABServer() *MATLABServer {
	matlabPath := os.Getenv("MATLAB_PATH")
	if matlabPath == "" {
		matlabPath = "matlab" // Default to system PATH
	}
	return &MATLABServer{
		matlabPath: matlabPath,
	}
}

// executeMATLAB executes MATLAB code using matlab -batch command
func (s *MATLABServer) executeMATLAB(code string, timeout time.Duration) (string, error) {
	// Create a temporary file for the MATLAB code
	tmpFile, err := os.CreateTemp("", "matlab_code_*.m")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write the MATLAB code to the temporary file
	if _, err := tmpFile.WriteString(code); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write MATLAB code to file: %v", err)
	}
	tmpFile.Close()

	// Get the base name without extension for the -batch command
	baseName := strings.TrimSuffix(filepath.Base(tmpFile.Name()), ".m")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Execute MATLAB with -batch option
	cmd := exec.CommandContext(ctx, s.matlabPath, "-batch", baseName)
	cmd.Dir = filepath.Dir(tmpFile.Name())

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("MATLAB execution timed out after %v", timeout)
	}

	if err != nil {
		return string(output), fmt.Errorf("MATLAB execution failed: %v", err)
	}

	return string(output), nil
}

// handleExecuteMATLAB handles the execute_matlab tool call
func (s *MATLABServer) handleExecuteMATLAB(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract code from arguments using the correct API
	code, err := request.RequireString("code")
	if err != nil {
		return nil, fmt.Errorf("code parameter is required: %v", err)
	}

	// Get timeout, default to 300 seconds (5 minutes)
	timeout := time.Duration(request.GetFloat("timeout", 300)) * time.Second

	// Execute MATLAB code
	output, err := s.executeMATLAB(code, timeout)

	if err != nil {
		// Return error result using the helper function
		return mcp.NewToolResultError(fmt.Sprintf("Error executing MATLAB code: %v\nOutput: %s", err, output)), nil
	}

	// Return successful result using the helper function
	return mcp.NewToolResultText(output), nil
}

func main() {
	matlabServer := NewMATLABServer()

	log.Printf("Starting MATLAB Batch MCP Server")
	log.Printf("MATLAB Path: %s", matlabServer.matlabPath)

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"batch-mcp",
		"1.0.0",
	)

	// Create the tool using the builder pattern
	tool := mcp.NewTool("execute_matlab",
		mcp.WithDescription("Execute MATLAB code using matlab -batch command"),
		mcp.WithString("code", mcp.Required(), mcp.Description("MATLAB code to execute")),
		mcp.WithNumber("timeout", mcp.Description("Timeout in seconds (default: 300)"), mcp.DefaultNumber(300)),
	)

	// Register the execute_matlab tool
	mcpServer.AddTool(tool, matlabServer.handleExecuteMATLAB)

	var transport string
	flag.StringVar(&transport, "t", "stdio", "Transport type (stdio or http)")
	flag.StringVar(&transport, "transport", "stdio", "Transport type (stdio or http)")
	flag.Parse()

	// Only check for "http" since stdio is the default
	if transport == "http" {
		httpServer := server.NewStreamableHTTPServer(mcpServer)
		log.Printf("HTTP server listening on :8080/mcp")
		if err := httpServer.Start(":8080"); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}
