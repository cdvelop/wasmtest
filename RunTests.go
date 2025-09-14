package wasmtest

import (
	"context"
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

	// Check if directory exists before attempting to change
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("test directory %s does not exist", dir)
	}

	// Change to the test directory
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("failed to change to directory %s: %v", dir, err)
	}

	// Check for WebAssembly test files
	hasWasmTests := false
	files, err := os.ReadDir(".")
	if err == nil {
		for _, file := range files {
			if strings.HasSuffix(file.Name(), "_test.go") {
				content, err := os.ReadFile(file.Name())
				if err == nil && (strings.Contains(string(content), "//go:build js && wasm") ||
					strings.Contains(string(content), "// +build js,wasm")) {
					hasWasmTests = true
					break
				}
			}
		}
	}

	if !hasWasmTests {
		return fmt.Errorf("no WebAssembly test files found in directory %s (files must contain '//go:build js && wasm' or '// +build js,wasm')", dir)
	}

	// Create Wasmtest instance
	w := New(logger)

	// Collect progress messages to determine success/failure
	var messages [][]any
	var lastMessage []any
	var hasErrors bool
	var errorMessages []string

	progressFunc := func(msgs ...any) {
		messages = append(messages, msgs)
		lastMessage = msgs

		// Log all messages if logger is provided
		if logger != nil {
			logger(append([]any{"[WASMTEST]"}, msgs...)...)
		}

		// Check for errors
		if len(msgs) > 0 {
			msgType := fmt.Sprintf("%v", msgs[0])
			if msgType == "error" || msgType == "err" {
				hasErrors = true
				if len(msgs) > 1 {
					errorMessages = append(errorMessages, fmt.Sprintf("%v", msgs[1]))
				}
			}
		}
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
		return fmt.Errorf("test execution timed out after %v in directory %s", timeout, dir)
	}

	// Analyze results from progress messages
	if len(messages) == 0 {
		return fmt.Errorf("no progress messages received - possible test execution failure in directory %s", dir)
	}

	// Check for errors in output
	if hasErrors {
		errorSummary := strings.Join(errorMessages, "; ")
		if errorSummary == "" {
			errorSummary = "unknown error occurred during test execution"
		}
		return fmt.Errorf("WebAssembly test execution failed in directory %s: %s", dir, errorSummary)
	}

	if len(lastMessage) >= 2 {
		if lastMessage[0] == "exit" {
			if lastMessage[1] == "error" {
				errorDetail := "unknown error"
				if len(lastMessage) > 2 {
					errorDetail = fmt.Sprintf("%v", lastMessage[2])
				}
				return fmt.Errorf("WebAssembly tests failed in directory %s: %s", dir, errorDetail)
			} else if lastMessage[1] == "ok" {
				// Check for PASS in output to confirm success
				foundPass := false
				for _, msg := range messages {
					if len(msg) >= 2 && msg[0] == "out" {
						output := fmt.Sprintf("%v", msg[1])
						if strings.Contains(output, "PASS") {
							foundPass = true
							break
						}
					}
				}
				if foundPass {
					return nil // Success
				}
				return fmt.Errorf("tests completed in directory %s but no PASS found in output - check test implementation", dir)
			}
		}
	}

	// If we reach here, something unexpected happened
	var debugInfo strings.Builder
	debugInfo.WriteString(fmt.Sprintf("unexpected test execution result in directory %s\n", dir))
	debugInfo.WriteString(fmt.Sprintf("Total messages received: %d\n", len(messages)))
	if len(lastMessage) > 0 {
		debugInfo.WriteString(fmt.Sprintf("Last message: %v\n", lastMessage))
	}
	debugInfo.WriteString("All messages:\n")
	for i, msg := range messages {
		debugInfo.WriteString(fmt.Sprintf("  %d: %v\n", i+1, msg))
	}

	return fmt.Errorf("WebAssembly test execution issue in directory %s:\n%s", dir, debugInfo.String())
}
