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

### Running Integration Tests

From [integration_test.go](integration_test.go), WasmTest executes real WASM tests in the `example/` directory:

1. Change to a directory with WASM-compatible tests (e.g., `example/` with `go.mod` and test files).
2. Set `GOOS=js GOARCH=wasm`.
3. Use `Execute` as above; it compiles the test binary, serves via HTTP, loads in Chrome (headless), and reports via progress.

Expected flow (from test logs):
- Compiles `go.test` to WASM.
- Starts local server.
- Launches browser, runs tests.
- Captures output: e.g., "PASS" for success.
- Exits with status.

### WASM Test Examples (from example/dom_test.go)

Your WASM tests can include:
- **Math Operations** (pure Go/WASM):
  ```go
  //go:build js && wasm

  package example

  type MathHelper struct{}

  func (m *MathHelper) Add(a, b int) int { return a + b }
  func (m *MathHelper) Multiply(a, b int) int { return a * b }
  func (m *MathHelper) Factorial(n int) int {
  	if n <= 1 { return 1 }
  	return n * m.Factorial(n-1)
  }

  func TestMathHelper(t *testing.T) {
  	m := &MathHelper{}
  	if m.Add(2, 3) != 5 { t.Error("Add failed") }
  	if m.Multiply(4, 5) != 20 { t.Error("Multiply failed") }
  	if m.Factorial(5) != 120 { t.Error("Factorial failed") }
  }
  ```

- **DOM Manipulation** (browser-specific):
  ```go
  //go:build js && wasm

  package example

  import (
  	"syscall/js"
  	"testing"
  )

  type DOMHelper struct{}

  func (d *DOMHelper) CreateElement(tag string) js.Value {
  	return js.Global().Get("document").Call("createElement", tag)
  }

  func (d *DOMHelper) SetInnerHTML(el js.Value, html string) {
  	el.Set("innerHTML", html)
  }

  func TestDOMHelper(t *testing.T) {
  	if js.Global().Get("document").IsUndefined() { t.Skip("Not in browser") }
  	dom := &DOMHelper{}
  	div := dom.CreateElement("div")
  	dom.SetInnerHTML(div, "Hello, WASM!")
  	if div.Get("innerHTML").String() != "Hello, WASM!" { t.Error("SetInnerHTML failed") }
  }
  ```

- **JS Interop**:
  ```go
  func TestJSInterop(t *testing.T) {
  	global := js.Global()
  	obj := global.Get("Object").New()
  	obj.Set("testProp", "testValue")
  	if obj.Get("testProp").String() != "testValue" { t.Error("JS prop failed") }
  }
  ```

Benchmarks like [`BenchmarkMathOperations`](example/dom_test.go:74) also work.

### Demo Command

Build and run the demo in `cmd/wasmtest-demo/`:
```
cd cmd/wasmtest-demo
go build -o wasmtest-demo .
./wasmtest-demo
```
It demonstrates basic execution (may require test files; see [integration_test.go](integration_test.go) for full flow).

## Advanced Usage

### CPU Profiling

Pass `-cpuprofile profile.pprof` to `go test` before calling `Execute`. WasmTest captures and converts profiles to Go's pprof format for analysis with `go tool pprof`.

### Running Non-Test Code

For `go run` (e.g., WASM apps): Set `WASM_HEADLESS=off` in env to view in browser. WasmTest supports this via the underlying tool.

### Coverage

For Go 1.20+: Use `-test.gocoverdir=/path/to/coverage` instead of `-test.coverprofile` to avoid large HTTP transfers. Post-process with `go tool covdata -i /path/to/coverage -o coverage.out`. Multiple runs can merge data.

### CI Integration (Travis/Github Actions)

WasmTest auto-installs wasmbrowsertest, but for CI:

- **Travis CI** (.travis.yml):
  ```
  addons:
    chrome: stable
  install:
  - go install github.com/agnivade/wasmbrowsertest@latest
  - export PATH=$(go env GOPATH)/bin:$PATH
  script:
  - GOOS=js GOARCH=wasm go test ./...
  ```

- **Github Actions** (.github/workflows/ci.yml):
  ```
  jobs:
    test:
      runs-on: ubuntu-latest
      steps:
      - uses: actions/setup-go@v2
        with: { go-version: '1.20' }
      - uses: browser-actions/setup-chrome@latest
      - run: go install github.com/agnivade/wasmbrowsertest@latest
      - run: export PATH=$(go env GOPATH)/bin:$PATH
      - uses: actions/checkout@v2
      - run: GOOS=js GOARCH=wasm go test ./...
  ```

Adjust Go version as needed. For custom setups, ensure Chrome (Blink-based) is installed.

### Browser Support

Uses Chrome DevTools Protocol, so supports Chrome and other Blink-based browsers (e.g., Edge). Firefox not supported due to geckodriver limitations.

### Custom Progress Handling

For TUIs or advanced logging, implement progress to update UI (e.g., show real-time output). See [tui.go](tui.go) for an example TUI integration.

## Troubleshooting

### Installation Fails

- Ensure `go` is in PATH and internet access for `go install`.
- Check logs from [`New`](wasmtest.go:19) for errors (e.g., timeout after 5min).
- Manually install: `go install github.com/agnivade/wasmbrowsertest@latest` and add to PATH.

### "total length of command line and environment variables exceeds limit"

Env vars (e.g., GITHUB_ in CI) too large. Use [cleanenv](https://github.com/agnivade/wasmbrowsertest/tree/main/cmd/cleanenv) to filter:
```
go install github.com/agnivade/wasmbrowsertest/cmd/cleanenv@latest
GOOS=js GOARCH=wasm cleanenv -remove-prefix GITHUB_ -- go test ./...
```

### No Progress Messages or Tests Skip

- Ensure tests have `//go:build js && wasm`.
- Verify in browser env (progress skips if no document for DOM tests).
- Check for `go.test` compilation errors in logs.

### Tests Fail in Headless Mode

Some DOM/JS tests may need `WASM_HEADLESS=off` for visual debugging.

For more, see underlying [wasmbrowsertest docs](docs/wasmbrowsertest.md).

## License

[MIT](LICENSE).
