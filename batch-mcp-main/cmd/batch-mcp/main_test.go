package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestMATLABServer_NewMATLABServer(t *testing.T) {
	tests := []struct {
		name         string
		envVar       string
		expectedPath string
	}{
		{
			name:         "with custom MATLAB path",
			envVar:       "/Applications/MATLAB_R2025a.app/bin/matlab",
			expectedPath: "/Applications/MATLAB_R2025a.app/bin/matlab",
		},
		{
			name:         "with default path",
			envVar:       "",
			expectedPath: "matlab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envVar != "" {
				os.Setenv("MATLAB_PATH", tt.envVar)
			} else {
				os.Unsetenv("MATLAB_PATH")
			}
			defer os.Unsetenv("MATLAB_PATH")

			server := NewMATLABServer()
			if server.matlabPath != tt.expectedPath {
				t.Errorf("NewMATLABServer() matlabPath = %v, want %v", server.matlabPath, tt.expectedPath)
			}
		})
	}
}

func TestMATLABServer_executeMATLAB(t *testing.T) {
	// Set the MATLAB path for testing
	os.Setenv("MATLAB_PATH", "/Applications/MATLAB_R2025a.app/bin/matlab")
	defer os.Unsetenv("MATLAB_PATH")

	server := NewMATLABServer()

	tests := []struct {
		name        string
		code        string
		timeout     time.Duration
		expectError bool
		contains    string
	}{
		{
			name:        "simple calculation",
			code:        "disp(2 + 3)",
			timeout:     30 * time.Second,
			expectError: false,
			contains:    "5",
		},
		{
			name:        "hello world",
			code:        "disp('Hello from MATLAB!')",
			timeout:     30 * time.Second,
			expectError: false,
			contains:    "Hello from MATLAB!",
		},
		{
			name:        "matrix operation",
			code:        "A = [1 2; 3 4]; disp(det(A))",
			timeout:     30 * time.Second,
			expectError: false,
			contains:    "-2",
		},
		{
			name:        "multiple operations",
			code:        "x = 1:5; y = x.^2; disp(sum(y))",
			timeout:     30 * time.Second,
			expectError: false,
			contains:    "55",
		},
		{
			name:        "timeout test",
			code:        "pause(2); disp('done')",
			timeout:     1 * time.Second,
			expectError: true,
			contains:    "timeout",
		},
		{
			name:        "syntax error",
			code:        "invalid matlab syntax +++",
			timeout:     30 * time.Second,
			expectError: true,
			contains:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := server.executeMATLAB(tt.code, tt.timeout)

			if tt.expectError && err == nil {
				t.Errorf("executeMATLAB() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("executeMATLAB() unexpected error: %v", err)
			}

			if tt.contains != "" && !strings.Contains(output, tt.contains) {
				t.Errorf("executeMATLAB() output = %v, expected to contain %v", output, tt.contains)
			}

			t.Logf("Code: %s", tt.code)
			t.Logf("Output: %s", output)
			if err != nil {
				t.Logf("Error: %v", err)
			}
		})
	}
}

func TestMATLABServer_handleExecuteMATLAB(t *testing.T) {
	// Set the MATLAB path for testing
	os.Setenv("MATLAB_PATH", "/Applications/MATLAB_R2025a.app/bin/matlab")
	defer os.Unsetenv("MATLAB_PATH")

	server := NewMATLABServer()
	ctx := context.Background()

	tests := []struct {
		name        string
		arguments   map[string]interface{}
		expectError bool
		contains    string
	}{
		{
			name: "valid request with code only",
			arguments: map[string]interface{}{
				"code": "disp('Test successful')",
			},
			expectError: false,
			contains:    "Test successful",
		},
		{
			name: "valid request with timeout",
			arguments: map[string]interface{}{
				"code":    "disp(42)",
				"timeout": 10.0,
			},
			expectError: false,
			contains:    "42",
		},
		{
			name: "missing code parameter",
			arguments: map[string]interface{}{
				"timeout": 30.0,
			},
			expectError: true,
			contains:    "",
		},
		{
			name: "invalid code type",
			arguments: map[string]interface{}{
				"code": 123,
			},
			expectError: true,
			contains:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock CallToolRequest
			request := createMockCallToolRequest(tt.arguments)

			result, err := server.handleExecuteMATLAB(ctx, request)

			if tt.expectError && err == nil {
				t.Errorf("handleExecuteMATLAB() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("handleExecuteMATLAB() unexpected error: %v", err)
			}

			if result != nil && tt.contains != "" {
				found := false
				for _, content := range result.Content {
					if textContent, ok := content.(mcp.TextContent); ok {
						if strings.Contains(textContent.Text, tt.contains) {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("handleExecuteMATLAB() result content does not contain %v", tt.contains)
				}
			}

			if result != nil {
				t.Logf("Result IsError: %v", result.IsError)
				for i, content := range result.Content {
					if textContent, ok := content.(mcp.TextContent); ok {
						t.Logf("Content[%d]: %s", i, textContent.Text)
					}
				}
			}
		})
	}
}

// Helper function to create a mock CallToolRequest
func createMockCallToolRequest(arguments map[string]interface{}) mcp.CallToolRequest {
	// Create the request structure manually since we need to test the handler
	params := struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}{
		Name:      "execute_matlab",
		Arguments: arguments,
	}

	// Convert to JSON and back to simulate real request
	jsonData, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params":  params,
	})

	var request mcp.CallToolRequest
	json.Unmarshal(jsonData, &request)

	return request
}

// Integration test that requires MATLAB to be installed
func TestMATLABIntegration(t *testing.T) {
	// Skip if MATLAB_PATH is not set or MATLAB is not available
	matlabPath := "/Applications/MATLAB_R2025a.app/bin/matlab"
	if _, err := os.Stat(matlabPath); os.IsNotExist(err) {
		t.Skipf("MATLAB not found at %s, skipping integration test", matlabPath)
	}

	os.Setenv("MATLAB_PATH", matlabPath)
	defer os.Unsetenv("MATLAB_PATH")

	server := NewMATLABServer()

	// Test basic MATLAB functionality
	testCases := []struct {
		name string
		code string
		want string
	}{
		{
			name: "basic arithmetic",
			code: "result = 10 + 5; disp(result)",
			want: "15",
		},
		{
			name: "string output",
			code: "disp('MATLAB is working!')",
			want: "MATLAB is working!",
		},
		{
			name: "matrix operations",
			code: "A = eye(3); disp(trace(A))",
			want: "3",
		},
		{
			name: "mathematical functions",
			code: "x = pi/4; disp(sin(x))",
			want: "0.7071", // approximately sqrt(2)/2
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := server.executeMATLAB(tc.code, 30*time.Second)
			if err != nil {
				t.Fatalf("executeMATLAB failed: %v", err)
			}

			if !strings.Contains(output, tc.want) {
				t.Errorf("Expected output to contain %q, got: %s", tc.want, output)
			}

			t.Logf("Code: %s", tc.code)
			t.Logf("Output: %s", output)
		})
	}
}

// Benchmark test for MATLAB execution
func BenchmarkMATLABExecution(b *testing.B) {
	// Skip if MATLAB is not available
	matlabPath := "/Applications/MATLAB_R2025a.app/bin/matlab"
	if _, err := os.Stat(matlabPath); os.IsNotExist(err) {
		b.Skipf("MATLAB not found at %s, skipping benchmark", matlabPath)
	}

	os.Setenv("MATLAB_PATH", matlabPath)
	defer os.Unsetenv("MATLAB_PATH")

	server := NewMATLABServer()
	code := "disp(2 + 2)"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := server.executeMATLAB(code, 30*time.Second)
		if err != nil {
			b.Fatalf("executeMATLAB failed: %v", err)
		}
	}
}
