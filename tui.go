package wasmtest

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Name returns an identifier for the handler.
func (w *Wasmtest) Name() string { return "WasmTest" }

// Label returns a user-friendly label for the handler (for UI buttons).
func (w *Wasmtest) Label() string { return "Ensure Wasm Browser Test" }

// Execute implements the HandlerExecution interface. It runs the installation
// procedure and reports progress through the provided callback.
func (w *Wasmtest) Execute(progress func(msgs ...any)) {
	if progress == nil {
		// nothing to report
		return
	}

	// Ensure go_js_wasm_exec is available for WASM test execution
	if err := w.ensureWasmExecSymlink(progress); err != nil {
		progress("error", "failed to setup WASM executor:", err)
		return
	}

	// create a background context with a short timeout for UI operations
	ctx, cancel := contextWithTimeout(10 * time.Minute)
	defer cancel()

	// Run the documented command: GOOS=js GOARCH=wasm go test -v
	cmd := exec.CommandContext(ctx, "go", "test", "-v")
	// Set environment variables for the command copied from the parent's env
	env := os.Environ()
	// ensure GOOS and GOARCH are set to js/wasm
	env = append(env, "GOOS=js", "GOARCH=wasm")
	cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		progress("error", "stdout pipe error:", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		progress("error", "stderr pipe error:", err)
		return
	}

	if err := cmd.Start(); err != nil {
		progress("error", "failed to start go test:", err)
		return
	}

	// stream stdout and stderr lines to progress
	stream := func(r *bufio.Reader, tag string) {
		for {
			line, err := r.ReadString('\n')
			if line != "" {
				progress(tag, strings.TrimRight(line, "\n"))
			}
			if err != nil {
				return
			}
		}
	}

	go stream(bufio.NewReader(stdout), "out")
	go stream(bufio.NewReader(stderr), "err")

	if err := cmd.Wait(); err != nil {
		progress("exit", "error", err.Error())
		return
	}
	progress("exit", "ok")
}

// GetLastOperationID implements MessageTracker.
func (w *Wasmtest) GetLastOperationID() string {
	return w.lastOpID
}

// SetLastOperationID implements MessageTracker.
func (w *Wasmtest) SetLastOperationID(id string) {
	w.lastOpID = id
}

// contextWithTimeout creates a context with timeout. Placed here to avoid
// importing context in this small file directly (wasmtest.go already imports it).
func contextWithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}

// ensureWasmExecSymlink ensures that go_js_wasm_exec exists for WASM test execution.
// It checks if go_js_wasm_exec exists, and if not, creates a symlink to wasmbrowsertest.
func (w *Wasmtest) ensureWasmExecSymlink(progress func(msgs ...any)) error {
	// First check if go_js_wasm_exec already exists
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		if cmd := exec.Command("go", "env", "GOPATH"); cmd != nil {
			if output, err := cmd.Output(); err == nil {
				gopath = strings.TrimSpace(string(output))
			}
		}
	}

	if gopath == "" {
		return nil // Skip if we can't determine GOPATH
	}

	goWasmExec := gopath + "/bin/go_js_wasm_exec"
	wasmBrowserTest := gopath + "/bin/wasmbrowsertest"

	// Check if go_js_wasm_exec already exists
	if _, err := os.Stat(goWasmExec); err == nil {
		progress("info", "go_js_wasm_exec already exists")
		return nil
	}

	// Check if wasmbrowsertest exists
	if _, err := os.Stat(wasmBrowserTest); os.IsNotExist(err) {
		if w.log != nil {
			w.log("wasmbrowsertest not found, automatic installation may be in progress")
		}
		progress("warning", "wasmbrowsertest not found, tests may fail if not installed")
		return nil // Don't fail, let the test try anyway
	}

	// Create symlink from wasmbrowsertest to go_js_wasm_exec
	if err := os.Symlink(wasmBrowserTest, goWasmExec); err != nil {
		progress("warning", "failed to create go_js_wasm_exec symlink:", err.Error())
		return nil // Don't fail, let the test try anyway
	}

	progress("info", "created go_js_wasm_exec -> wasmbrowsertest symlink")
	return nil
}
