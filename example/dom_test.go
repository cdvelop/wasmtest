//go:build js && wasm
// +build js,wasm

package example

import (
	"syscall/js"
	"testing"
)

func TestMathHelper(t *testing.T) {
	m := &MathHelper{}

	// Test Add
	result := m.Add(2, 3)
	if result != 5 {
		t.Errorf("Add(2, 3) = %d; want 5", result)
	}

	// Test Multiply
	result = m.Multiply(4, 5)
	if result != 20 {
		t.Errorf("Multiply(4, 5) = %d; want 20", result)
	}

	// Test Factorial
	result = m.Factorial(5)
	if result != 120 {
		t.Errorf("Factorial(5) = %d; want 120", result)
	}
}

func TestDOMHelper(t *testing.T) {
	// Skip if not in browser environment
	if js.Global().Get("document").IsUndefined() {
		t.Skip("Skipping DOM tests - not in browser environment")
	}

	dom := NewDOMHelper()

	// Test CreateElement
	div := dom.CreateElement("div")
	if div.IsNull() || div.IsUndefined() {
		t.Error("CreateElement should return a valid element")
	}

	// Test SetInnerHTML
	dom.SetInnerHTML(div, "Hello, WebAssembly!")
	innerHTML := div.Get("innerHTML").String()
	if innerHTML != "Hello, WebAssembly!" {
		t.Errorf("SetInnerHTML failed: got %q, want %q", innerHTML, "Hello, WebAssembly!")
	}
}

func TestJSInterop(t *testing.T) {
	// Test basic JS interaction
	global := js.Global()

	// Test that we can access global objects
	if global.Get("Object").IsUndefined() {
		t.Error("Should be able to access global Object")
	}

	// Test creating a JS object
	obj := global.Get("Object").New()
	obj.Set("testProp", "testValue")

	value := obj.Get("testProp").String()
	if value != "testValue" {
		t.Errorf("JS object property: got %q, want %q", value, "testValue")
	}
}

func BenchmarkMathOperations(b *testing.B) {
	m := &MathHelper{}
	for i := 0; i < b.N; i++ {
		_ = m.Add(i, i+1)
		_ = m.Multiply(i, 2)
	}
}
