# Examples

## Running Integration Tests

From [`integration_test.go`](integration_test.go), WasmTest executes real WASM tests in the `example/` directory:

1. Change to a directory with WASM-compatible tests (e.g., `example/` with `go.mod` and test files).
2. Set `GOOS=js GOARCH=wasm`.
3. Use [`Execute`](wasmtest.go) as in basic usage; it compiles the test binary, serves via HTTP, loads in Chrome (headless), and reports via progress.

Expected flow (from test logs):
- Compiles `go.test` to WASM.
- Starts local server.
- Launches browser, runs tests.
- Captures output: e.g., "PASS" for success.
- Exits with status.

## WASM Test Examples (from example/dom_test.go)

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

## Demo Command

Build and run the demo in `cmd/wasmtest-demo/`:
```
cd cmd/wasmtest-demo
go build -o wasmtest-demo .
./wasmtest-demo
```
It demonstrates basic execution (may require test files; see [`integration_test.go`](integration_test.go) for full flow).