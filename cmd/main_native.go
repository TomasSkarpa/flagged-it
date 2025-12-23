//go:build !js || !wasm
// +build !js !wasm

package main

import (
	"fyne.io/fyne/v2"
)

// setupVisibilityHandler is a no-op for native builds
func setupVisibilityHandler(window fyne.Window) {
	// Not needed for native builds
}

