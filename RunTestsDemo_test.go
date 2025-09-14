package wasmtest_test

import (
	. "github.com/cdvelop/wasmtest"
	"testing"
	"time"
)

// RunTestsDemo tests the simplified RunTests API by running tests in the example directory
func RunTestsDemo(t *testing.T) {
	// Test the simplified RunTests API - ultra simple!
	if err := RunTests("./example", nil, 10*time.Minute); err != nil {
		t.Errorf("RunTests failed: %v", err)
	}
}
