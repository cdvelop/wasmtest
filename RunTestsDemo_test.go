package wasmtest

import (
	"testing"
	"time"
)

// TestRunTestsDemo tests the simplified RunTests API by running tests in the example directory
func TestRunTestsDemo(t *testing.T) {
	// Test the simplified RunTests API - ultra simple!
	if err := RunTests("./example", func(a ...any) { t.Log(a) }, 10*time.Minute); err != nil {
		t.Errorf("RunTests failed: %v", err)
	}
}
