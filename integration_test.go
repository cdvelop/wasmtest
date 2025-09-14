package wasmtest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestExecuteIntegration tests the complete flow of Execute running real WASM tests
func TestExecuteIntegration(t *testing.T) {
	// Skip if 'go' not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not found in PATH; skipping integration test")
	}

	// Ensure we're in the right directory structure
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	exampleDir := filepath.Join(cwd, "example")
	if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
		t.Skipf("Example directory not found: %s", exampleDir)
	}

	// Change to example directory
	originalDir := cwd
	defer os.Chdir(originalDir)

	if err := os.Chdir(exampleDir); err != nil {
		t.Fatalf("Failed to change to example directory: %v", err)
	}

	// Collect progress messages
	var messages [][]any
	var lastMessage []any

	progressFunc := func(msgs ...any) {
		messages = append(messages, msgs)
		lastMessage = msgs
		t.Logf("Progress: %v", msgs)
	}

	// Create logger
	logger := func(msgs ...any) {
		t.Logf("Log: %v", msgs)
	}

	// Initialize Wasmtest
	w := New(logger)

	// Test the interfaces
	if w.Name() == "" {
		t.Error("Name() should return non-empty string")
	}
	if w.Label() == "" {
		t.Error("Label() should return non-empty string")
	}

	// Test MessageTracker interface
	testID := "test-op-123"
	w.SetLastOperationID(testID)
	if w.GetLastOperationID() != testID {
		t.Errorf("GetLastOperationID() = %q, want %q", w.GetLastOperationID(), testID)
	}

	// Execute the WASM tests
	t.Log("Starting Execute...")
	start := time.Now()
	w.Execute(progressFunc)
	duration := time.Since(start)

	t.Logf("Execute completed in %v", duration)
	t.Logf("Total progress messages: %d", len(messages))

	// Verify we got some progress messages
	if len(messages) == 0 {
		t.Error("Expected progress messages, got none")
	}

	// Check if we got an exit message
	hasExit := false
	for _, msg := range messages {
		if len(msg) > 0 {
			if str, ok := msg[0].(string); ok && str == "exit" {
				hasExit = true
				break
			}
		}
	}

	if !hasExit {
		t.Error("Expected 'exit' message in progress")
	}

	// Log all messages for debugging
	t.Log("All progress messages:")
	for i, msg := range messages {
		t.Logf("  %d: %v", i, msg)
	}

	// Check the last message to see if it indicates success or failure
	if len(lastMessage) >= 2 {
		if lastMessage[0] == "exit" {
			if lastMessage[1] == "error" {
				t.Errorf("WASM tests failed with error: %v", lastMessage)
			} else if lastMessage[1] == "ok" {
				t.Log("WASM tests completed successfully!")

				// Additional validation: check for PASS in output
				hasPass := false
				for _, msg := range messages {
					if len(msg) >= 2 && msg[0] == "out" {
						if strings.Contains(fmt.Sprintf("%v", msg[1]), "PASS") {
							hasPass = true
							break
						}
					}
				}

				if !hasPass {
					t.Error("Expected PASS in test output but didn't find it")
				}
			} else {
				t.Errorf("Unexpected exit status: %v", lastMessage)
			}
		}
	} else {
		t.Error("Expected exit message with status but didn't receive one")
	}
}

// TestExecuteWithNilProgress tests Execute with nil progress function
func TestExecuteWithNilProgress(t *testing.T) {
	logger := func(msgs ...any) {
		t.Logf("Log: %v", msgs)
	}

	w := New(logger)

	// This should not panic and should return immediately
	w.Execute(nil)

	t.Log("Execute with nil progress completed without panic")
}
