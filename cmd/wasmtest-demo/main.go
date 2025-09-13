package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/cdvelop/wasmtest"
)

func main() {
	fmt.Println("=== WasmTest Demo ===")

	// Create a logger that will capture progress messages
	var messages []string
	logger := func(msgs ...any) {
		for _, msg := range msgs {
			fmt.Printf("[LOG] %v\n", msg)
		}
	}

	progressLogger := func(msgs ...any) {
		message := fmt.Sprintf("%v", msgs)
		messages = append(messages, message)
		fmt.Printf("[PROGRESS] %s\n", message)
	}

	// Initialize Wasmtest
	w := wasmtest.New(logger)

	fmt.Printf("Handler Name: %s\n", w.Name())
	fmt.Printf("Handler Label: %s\n", w.Label())

	// Change to the example directory to run tests there
	exampleDir := filepath.Join("..", "..", "example")
	if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
		log.Fatalf("Example directory not found: %s", exampleDir)
	}

	// Change working directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(exampleDir); err != nil {
		log.Fatalf("Failed to change to example directory: %v", err)
	}

	fmt.Printf("Changed to directory: %s\n", exampleDir)
	fmt.Println("Executing WASM tests...")

	// Execute the WASM tests
	w.Execute(progressLogger)

	fmt.Println("\n=== Test Execution Complete ===")
	fmt.Printf("Total progress messages: %d\n", len(messages))

	// Show some statistics
	if len(messages) > 0 {
		fmt.Println("First few messages:")
		for i, msg := range messages {
			if i >= 5 { // Show only first 5
				break
			}
			fmt.Printf("  %d: %s\n", i+1, msg)
		}
	}
}
