package wasmtest

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"
)

// RunTests provides a simplified variadic API for running WebAssembly tests.
// It accepts optional arguments of types: string (directory), func(...any) (logger), time.Duration (timeout).
// Defaults: dir="wasm_tests", logger=fmt.Println, timeout=3*time.Minute
//
// Examples:
//
//	RunTests()                          // uses all defaults
//	RunTests("./my_tests")              // sets custom directory
//	RunTests(myLogger)                  // sets custom logger
//	RunTests(5 * time.Minute)           // sets custom timeout
//	RunTests("./my_tests", myLogger)    // sets directory and logger
//
// Note: if dir is passed as an empty string "" or ".", it defaults to "wasm_tests".
func RunTests(args ...any) error {
	// Parse variadic arguments by type
	dir := "wasm_tests"
	logger := func(a ...any) { fmt.Println(a...) }
	timeout := 3 * time.Minute
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			dir = v
		case func(...any):
			logger = v
		case time.Duration:
			timeout = v
		}
	}
	// Get current directory to restore later
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("❌💥 CRITICAL ERROR: Failed to get current directory\n🔴 Details: %v", err)
	}
	defer os.Chdir(originalDir)

	// Normalize dir: if empty or "." use "wasm_tests"
	if dir == "" || dir == "." {
		dir = "wasm_tests"
	}

	// Check if directory exists before attempting to change
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("❌💥 DIRECTORY ERROR: Test directory %s does not exist\n🔴 Please ensure the test directory exists and contains WebAssembly test files", dir)
	}

	// Change to the test directory
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("❌💥 DIRECTORY ERROR: Failed to change to directory %s\n🔴 Details: %v", dir, err)
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
		return fmt.Errorf("❌💥 NO TEST FILES: No WebAssembly test files found in directory %s\n🔴 Required: Files must contain '//go:build js && wasm' or '// +build js,wasm'\n💡 Check that your test files have the correct build tags", dir)
	}

	// Create Wasmtest instance
	w := New(logger)

	// Collect progress messages to determine success/failure
	var messages [][]any
	var lastMessage []any
	var hasErrors bool
	var errorMessages []string
	var failedTests []string

	progressFunc := func(msgs ...any) {
		messages = append(messages, msgs)
		lastMessage = msgs

		// Log all messages
		if len(msgs) > 1 && msgs[0] == "out" {
			// For test output lines, log without [WASMTEST] prefix for cleaner display
			logger(msgs[1:]...)
		} else {
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

			// Parse test output to find failing tests
			if msgType == "out" && len(msgs) > 1 {
				output := fmt.Sprintf("%v", msgs[1])
				// Look for test failure patterns
				if strings.Contains(output, "--- FAIL:") {
					// Extract test name from patterns like:
					// "--- FAIL: TestFunctionName (0.00s)"
					// "--- FAIL: TestFunctionName/SubTest (0.00s)"
					if strings.HasPrefix(output, "--- FAIL: ") {
						testName := strings.TrimPrefix(output, "--- FAIL: ")
						// Remove timing info
						if idx := strings.LastIndex(testName, " ("); idx > 0 {
							testName = testName[:idx]
						}
						if testName != "" && !slices.Contains(failedTests, testName) {
							failedTests = append(failedTests, testName)
						}
					}
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
		return fmt.Errorf("⏰💥 TIMEOUT ERROR: Test execution timed out after %v in directory %s\n🔴 This usually means the WebAssembly tests are hanging or taking too long\n💡 Try increasing the timeout or check for infinite loops in your tests", timeout, dir)
	}

	// Analyze results from progress messages
	if len(messages) == 0 {
		return fmt.Errorf("❌💥 NO OUTPUT: No progress messages received from test execution in directory %s\n🔴 This indicates a serious problem with the test runner\n💡 Check that wasmbrowsertest is properly installed and accessible", dir)
	}

	// Check for errors in output
	if hasErrors {
		errorSummary := strings.Join(errorMessages, "; ")
		if errorSummary == "" {
			errorSummary = "unknown error occurred during test execution"
		}

		errorMsg := fmt.Sprintf("❌💥 WebAssembly test EXECUTION FAILED in directory %s\n🔴 Error: %s", dir, errorSummary)

		// Add information about failing tests if any were found
		if len(failedTests) > 0 {
			errorMsg += fmt.Sprintf("\n🧪 Failing Tests: %s", strings.Join(failedTests, ", "))
		}

		errorMsg += "\n💡 Check the test output above for detailed failure information"
		return fmt.Errorf("%s", errorMsg)
	}

	if len(lastMessage) >= 2 {
		if lastMessage[0] == "exit" {
			if lastMessage[1] == "error" {
				errorDetail := "unknown error"
				if len(lastMessage) > 2 {
					errorDetail = fmt.Sprintf("%v", lastMessage[2])
				}

				errorMsg := fmt.Sprintf("❌💥 WebAssembly tests FAILED in directory %s\n🔴 Exit Error: %s", dir, errorDetail)

				// Add information about failing tests if any were found
				if len(failedTests) > 0 {
					errorMsg += fmt.Sprintf("\n🧪 Failing Tests: %s", strings.Join(failedTests, ", "))
				}

				errorMsg += "\n💡 The test process exited with an error status"
				return fmt.Errorf("%s", errorMsg)
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

				errorMsg := fmt.Sprintf("⚠️💥 PARTIAL SUCCESS: Tests completed in directory %s but no PASS found in output", dir)

				// Add information about failing tests if any were found
				if len(failedTests) > 0 {
					errorMsg += fmt.Sprintf("\n🧪 Failing Tests: %s", strings.Join(failedTests, ", "))
				}

				errorMsg += "\n🔴 This usually means your tests are not producing the expected output\n💡 Check that your test functions are named correctly (TestXxx) and contain proper assertions"
				return fmt.Errorf("%s", errorMsg)
			}
		}
	}

	// If we reach here, something unexpected happened
	var debugInfo strings.Builder
	debugInfo.WriteString(fmt.Sprintf("❌💥 UNEXPECTED ERROR: Test execution issue in directory %s\n", dir))
	debugInfo.WriteString(fmt.Sprintf("🔴 Total messages received: %d\n", len(messages)))
	if len(lastMessage) > 0 {
		debugInfo.WriteString(fmt.Sprintf("🔴 Last message: %v\n", lastMessage))
	}
	debugInfo.WriteString("📋 All messages received:\n")
	for i, msg := range messages {
		debugInfo.WriteString(fmt.Sprintf("  %d: %v\n", i+1, msg))
	}
	debugInfo.WriteString("💡 This usually indicates a problem with the test runner or environment setup")

	return fmt.Errorf("%s", debugInfo.String())
}
