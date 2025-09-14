package wasmtest

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// RunTests provides a simplified API for running WebAssembly tests.
// It changes to the specified directory, runs the tests, and returns an error if they fail.
// The logger parameter is optional (pass nil for no logging).
// The timeout specifies the maximum time to wait for tests to complete.
//
// Note about the `dir` parameter: if `dir` is passed as an empty string "" or as ".",
// it will default to the directory name "wasm_test". This makes it convenient to call
// RunTests("", ...) or RunTests(".", ...) when your WASM test files live under
// a `wasm_test` directory relative to the caller.
func RunTests(dir string, logger func(...any), timeout time.Duration) error {
	// Get current directory to restore later
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Normalize dir: if empty or "." use "wasm_test"
	if dir == "" || dir == "." {
		dir = "wasm_test"
	}

	// Change to the test directory
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("failed to change to directory %s: %v", dir, err)
	}

	// Create Wasmtest instance
	w := New(logger)

	// Collect progress messages to determine success/failure
	var messages [][]any
	var lastMessage []any

	progressFunc := func(msgs ...any) {
		messages = append(messages, msgs)
		lastMessage = msgs
	}

	// Execute tests with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Run Execute in a goroutine so we can timeout
	done := make(chan error, 1)
	go func() {
		// Note: Execute doesn't return error, it reports via progress callback
		w.Execute(progressFunc)
		done <- nil
	}()

	select {
	case <-done:
		// Execution completed
	case <-ctx.Done():
		return fmt.Errorf("test execution timed out after %v", timeout)
	}

	// Analyze results from progress messages
	if len(lastMessage) >= 2 {
		if lastMessage[0] == "exit" {
			if lastMessage[1] == "error" {
				return fmt.Errorf("WASM tests failed: %v", lastMessage[2:])
			} else if lastMessage[1] == "ok" {
				// Check for PASS in output to confirm success
				for _, msg := range messages {
					if len(msg) >= 2 && msg[0] == "out" {
						if strings.Contains(fmt.Sprintf("%v", msg[1]), "PASS") {
							return nil // Success
						}
					}
				}
				return errors.New("tests completed but no PASS found in output")
			}
		}
	}

	return errors.New("unexpected test execution result")
}
