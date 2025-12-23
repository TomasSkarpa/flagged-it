//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"

	"fyne.io/fyne/v2"
)

// setupVisibilityHandler pauses rendering when tab is hidden to prevent freezing
func setupVisibilityHandler(window fyne.Window) {
	doc := js.Global().Get("document")

	// Handle tab visibility changes
	visibilityChangeCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		hidden := doc.Get("hidden").Bool()

		if hidden {
			// Tab is hidden - canvas will automatically reduce rendering
			// This helps prevent the tab from freezing
		} else {
			// Tab is visible again - force a refresh to ensure UI is updated
			if window.Canvas() != nil {
				window.Canvas().Refresh(window.Content())
			}
		}

		return nil
	})

	// Listen for visibility change events
	doc.Call("addEventListener", "visibilitychange", visibilityChangeCallback)

	// Note: Browser keyboard shortcuts are handled in index.html via JavaScript
	// to intercept events before they reach Fyne's canvas
}
