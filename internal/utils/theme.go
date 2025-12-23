//go:build js && wasm
// +build js,wasm

package utils

import (
	"syscall/js"
)

// GetSavedTheme returns the saved theme preference from localStorage
// Returns "system", "light", or "dark"
func GetSavedTheme() string {
	localStorage := js.Global().Get("localStorage")
	if !localStorage.IsUndefined() {
		savedTheme := localStorage.Call("getItem", "theme")
		if !savedTheme.IsNull() && savedTheme.String() != "" {
			return savedTheme.String()
		}
	}
	return "system"
}

// SetSavedTheme saves the theme preference to localStorage
func SetSavedTheme(theme string) {
	localStorage := js.Global().Get("localStorage")
	if !localStorage.IsUndefined() {
		localStorage.Call("setItem", "theme", theme)
	}
}

// GetSystemTheme returns the system's preferred theme
func GetSystemTheme() string {
	window := js.Global().Get("window")
	if !window.IsUndefined() {
		matchMedia := window.Call("matchMedia", "(prefers-color-scheme: light)")
		if !matchMedia.IsUndefined() && matchMedia.Get("matches").Bool() {
			return "light"
		}
	}
	return "dark"
}

// GetEffectiveTheme returns the actual theme to use (resolves "system" to light/dark)
func GetEffectiveTheme() string {
	saved := GetSavedTheme()
	if saved == "system" {
		return GetSystemTheme()
	}
	return saved
}

// ReloadPage reloads the current page
func ReloadPage() {
	location := js.Global().Get("location")
	if !location.IsUndefined() {
		location.Call("reload")
	}
}

