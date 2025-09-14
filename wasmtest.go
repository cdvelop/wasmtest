package wasmtest

import (
	"context"
	"errors"
	"os/exec"
	"time"
)

type Wasmtest struct {
	// Log is a simple logger function provided by the caller.
	// It should accept variadic values similar to fmt.Println.
	log func(...any)
	// mutex to protect lastOpID
	lastOpID string
}

// New returns a Wasmtest configured with the provided logger.
// The logger must not be nil; if nil is passed a no-op logger is used.
func New(logger func(...any)) *Wasmtest {

	if logger == nil {
		logger = func(args ...any) {
			println(args)
		}
	}

	w := &Wasmtest{log: logger}

	// Perform a synchronous verification/install of wasmbrowsertest so callers
	// (and integrations like TUI) don't need to call it explicitly.
	go func() {
		// Run in background but wait a short period so New doesn't block too long.
		// Use a timeout context to bound the operation.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if err := w.ensureWasmBrowserTestInstalled(ctx); err != nil {
			// Log the error via provided logger
			w.log("ensure wasmbrowsertest failed:", err)
		}
	}()

	return w
}

// ensureWasmBrowserTestInstalled verifies that a binary named
// `wasmbrowsertest` (or `go_js_wasm_exec`) is available in PATH. If not
// present it will attempt to install `github.com/agnivade/wasmbrowsertest@latest`
// using `go install`. Errors are returned and also reported through the
// configured logger.
func (w *Wasmtest) ensureWasmBrowserTestInstalled(ctx context.Context) error {
	if w == nil {
		return errors.New("wasmtest: receiver is nil")
	}

	// Check common names: the original tool is `wasmbrowsertest` but the
	// README suggests renaming to `go_js_wasm_exec`. We'll accept either.
	probes := []string{"wasmbrowsertest", "go_js_wasm_exec"}

	for _, p := range probes {
		if _, err := exec.LookPath(p); err == nil {
			w.log("found", p)
			return nil
		}
	}

	w.log("wasmbrowsertest not found in PATH; attempting to install via go install")

	// Prepare install command with a timeout to avoid hanging indefinitely.
	// Use the module path from the docs.
	installCmd := exec.CommandContext(ctx, "go", "install", "github.com/agnivade/wasmbrowsertest@latest")
	// Set a reasonable timeout if the provided context has none.
	done := make(chan error, 1)
	go func() {
		done <- installCmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			w.log("go install failed:", err)
			return err
		}
	case <-ctx.Done():
		// kill process if still running
		_ = installCmd.Process.Kill()
		w.log("installation context cancelled or deadline exceeded:", ctx.Err())
		return ctx.Err()
	case <-time.After(2 * time.Minute):
		_ = installCmd.Process.Kill()
		w.log("installation timed out")
		return errors.New("wasmtest: go install timed out")
	}

	// After install, re-check PATH
	for _, p := range probes {
		if _, err := exec.LookPath(p); err == nil {
			w.log("installed and found", p)
			return nil
		}
	}

	w.log("installed but binary still not found in PATH; ensure GOBIN or GOPATH/bin is on PATH")
	return errors.New("wasmtest: installed but binary not found in PATH")
}
