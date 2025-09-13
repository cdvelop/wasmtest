package example

import (
	"syscall/js"
)

// DOMHelper provides utilities for DOM manipulation in WebAssembly
type DOMHelper struct {
	document js.Value
}

// NewDOMHelper creates a new DOM helper instance
func NewDOMHelper() *DOMHelper {
	return &DOMHelper{
		document: js.Global().Get("document"),
	}
}

// CreateElement creates a new HTML element with the given tag name
func (d *DOMHelper) CreateElement(tagName string) js.Value {
	return d.document.Call("createElement", tagName)
}

// GetElementByID returns an element by its ID
func (d *DOMHelper) GetElementByID(id string) js.Value {
	return d.document.Call("getElementById", id)
}

// SetInnerHTML sets the innerHTML of an element
func (d *DOMHelper) SetInnerHTML(element js.Value, html string) {
	element.Set("innerHTML", html)
}

// AddEventListener adds an event listener to an element
func (d *DOMHelper) AddEventListener(element js.Value, event string, callback js.Func) {
	element.Call("addEventListener", event, callback)
}

// MathHelper provides some math utilities that can be tested
type MathHelper struct{}

// Add adds two numbers
func (m *MathHelper) Add(a, b int) int {
	return a + b
}

// Multiply multiplies two numbers
func (m *MathHelper) Multiply(a, b int) int {
	return a * b
}

// Factorial calculates factorial of n
func (m *MathHelper) Factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * m.Factorial(n-1)
}
