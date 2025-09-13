# WasmTest — WebAssembly testing helper for Go

WasmTest is a Go library that simplifies running WebAssembly (WASM) tests for Go code targeting `js/wasm` in a browser environment. It wraps the popular [wasmbrowsertest](https://github.com/agnivade/wasmbrowsertest) tool, automating its installation and providing a clean API for execution with progress monitoring via callbacks. This makes it easier to integrate WASM testing into your Go projects, CI pipelines, or custom tools (e.g., TUIs).

Built for developers writing Go code that interacts with the browser DOM, JavaScript, or performs computations in WASM, WasmTest handles the boilerplate of compiling tests to WASM, serving them, and capturing output/errors from the browser.

## Installation

1. Add WasmTest to your Go module:
   ```
   go get github.com/cdvelop/wasmtest
   ```

2. Import in your code:
   ```go
   import "github.com/cdvelop/wasmtest"
   ```

3. WasmTest automatically ensures the underlying `wasmbrowsertest` binary (or `go_js_wasm_exec`) is installed via `go install github.com/agnivade/wasmbrowsertest@latest` when you create an instance with [`New`](wasmtest.go). No manual setup needed, though you can call [`EnsureWasmBrowserTestInstalled`](wasmtest.go:46) explicitly if desired.

   The binary will be placed in `$GOPATH/bin` or `$GOBIN`. Ensure this directory is in your `$PATH`.

## Basic Usage

Create a `Wasmtest` instance with a logger (optional, defaults to `println`), then call [`Execute`](wasmtest.go) with a progress callback to run tests. Set environment variables `GOOS=js GOARCH=wasm` before running tests.

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cdvelop/wasmtest"
)

func main() {
	// Simple logger (variadic like fmt.Println)
	logger := func(msgs ...any) {
		log.Println(msgs...)
	}

	// Create Wasmtest instance (auto-installs wasmbrowsertest in background)
	w := wasmtest.New(logger)

	// Progress callback: receives messages like ["out", "test output"], ["err", "error msg"], ["exit", "ok"|"error"]
	progress := func(msgs ...any) {
		if len(msgs) >= 2 && msgs[0] == "out" {
			fmt.Println("Test output:", msgs[1])
		} else if len(msgs) >= 2 && msgs[0] == "exit" {
			if msgs[1] == "ok" {
				fmt.Println("Tests passed!")
			} else {
				fmt.Printf("Tests failed: %v\n", msgs[1:])
				os.Exit(1)
			}
		}
	}

	// Run WASM tests (assumes current dir has go.test files)
	// Set GOOS=js GOARCH=wasm in env for cross-compilation
	if err := w.Execute(progress); err != nil {
		log.Fatal("Execution failed:", err)
	}
}
```

- [`New`](wasmtest.go:19)(logger): Initializes and starts background installation of wasmbrowsertest (non-blocking).
- [`Execute`](wasmtest.go)(progressFunc): Compiles and runs tests in browser, streaming progress via the callback. Blocks until completion.
- Progress messages: `["out", data]`, `["err", data]`, `["exit", "ok"|"error" [, details]]`.
- Use [`w.Name()`](wasmtest.go) and [`w.Label()`](wasmtest.go) for tool identification (e.g., in TUIs).

For full API details, see [wasmtest.go](wasmtest.go).

## Examples

See [docs/examples.md](docs/examples.md) for integration tests, WASM test code snippets (math, DOM, JS interop), and demo command.

## Advanced Usage

See [docs/advanced.md](docs/advanced.md) for CPU profiling, running non-test code, coverage, CI integration (Travis/Github Actions), browser support, and custom progress handling.

## Troubleshooting

See [docs/troubleshooting.md](docs/troubleshooting.md) for installation issues, environment variable limits, skipped tests, and headless mode problems.

For underlying tool details, see [docs/wasmbrowsertest.md](docs/wasmbrowsertest.md).

## License

[MIT](LICENSE).
