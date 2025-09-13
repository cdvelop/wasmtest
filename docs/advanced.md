# Advanced Usage

## CPU Profiling

Pass `-cpuprofile profile.pprof` to `go test` before calling [`Execute`](wasmtest.go). WasmTest captures and converts profiles to Go's pprof format for analysis with `go tool pprof`.

## Running Non-Test Code

For `go run` (e.g., WASM apps): Set `WASM_HEADLESS=off` in env to view in browser. WasmTest supports this via the underlying tool.

## Coverage

For Go 1.20+: Use `-test.gocoverdir=/path/to/coverage` instead of `-test.coverprofile` to avoid large HTTP transfers. Post-process with `go tool covdata -i /path/to/coverage -o coverage.out`. Multiple runs can merge data.

## CI Integration (Travis/Github Actions)

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

## Browser Support

Uses Chrome DevTools Protocol, so supports Chrome and other Blink-based browsers (e.g., Edge). Firefox not supported due to geckodriver limitations.

## Custom Progress Handling

For TUIs or advanced logging, implement progress to update UI (e.g., show real-time output). See [`tui.go`](tui.go) for an example TUI integration.