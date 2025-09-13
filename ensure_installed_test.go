package wasmtest

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestEnsureWasmBrowserTestInstalled_integration is an integration-style test
// that ensures EnsureWasmBrowserTestInstalled installs the binary when it's
// absent and leaves it installed at the end. It manipulates the user's PATH
// only by moving the binary out of the way and restoring it after the test.
func TestEnsureWasmBrowserTestInstalled_integration(t *testing.T) {
	// Ensure 'go' exists; if not, skip the test.
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not found in PATH; skipping integration test")
	}

	// candidate binary names
	probes := []string{"wasmbrowsertest", "go_js_wasm_exec"}

	// find existing binary and move it out of the way
	var foundPath string
	for _, p := range probes {
		if pth, err := exec.LookPath(p); err == nil {
			foundPath = pth
			break
		}
	}

	var backupPath string
	if foundPath != "" {
		// move it to a temporary location
		tmp := filepath.Join(os.TempDir(), "wasmbrowsertest.bak")
		if err := os.Rename(foundPath, tmp); err != nil {
			t.Fatalf("failed to move existing binary out of the way: %v", err)
		}
		backupPath = tmp
		t.Logf("moved existing binary %s -> %s", foundPath, backupPath)
	}

	// Ensure we restore at the end
	defer func() {
		if backupPath != "" {
			// if a binary with the original name exists now, remove it
			if cur, err := exec.LookPath(filepath.Base(foundPath)); err == nil {
				_ = os.Remove(cur)
			}
			// restore original
			if err := os.Rename(backupPath, foundPath); err != nil {
				t.Fatalf("failed to restore original binary: %v", err)
			}
			t.Logf("restored original binary to %s", foundPath)
		}
	}()

	// capture logs from our logger
	logs := []any{}
	logger := func(v ...any) {
		logs = append(logs, v)
	}

	// Create Wasmtest; New will perform the installation/verification in the
	// background. We wait here for a bit for it to complete.
	w := New(logger)

	// Use w to ensure it's referenced (and potentially to retrieve state).
	_ = w.GetLastOperationID()

	// Wait up to 6 minutes for the background installation to finish.
	waitUntil := time.Now().Add(6 * time.Minute)
	found := false
	for time.Now().Before(waitUntil) && !found {
		for _, p := range probes {
			if _, e := exec.LookPath(p); e == nil {
				found = true
				break
			}
		}
		if !found {
			time.Sleep(2 * time.Second)
		}
	}

	// After installation, ensure one of the probes exists in PATH
	ok := false
	for _, p := range probes {
		if _, err := exec.LookPath(p); err == nil {
			ok = true
			break
		}
	}
	if !ok {
		t.Fatalf("after EnsureWasmBrowserTestInstalled binary not found; logs=%v", logs)
	}
	t.Logf("installation succeeded; logs=%v", logs)
}
