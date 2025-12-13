//go:build js && wasm
// +build js,wasm

package utils

import (
	"os"
	"strings"
	"syscall/js"
)

func GetSystemLocale() string {
	// Try to get saved language from localStorage
	localStorage := js.Global().Get("localStorage")
	if !localStorage.IsUndefined() {
		savedLang := localStorage.Call("getItem", "language")
		if !savedLang.IsNull() && savedLang.String() != "" {
			return savedLang.String()
		}
	}

	// Fall back to browser language
	htmlLang := js.Global().Get("document").Get("documentElement").Get("lang")
	if !htmlLang.IsUndefined() && !htmlLang.IsNull() && htmlLang.String() != "" {
		return strings.Split(htmlLang.String(), "-")[0]
	}
	return "en"
}

func SetSystemLocale(locale string) {
	// Save to localStorage
	localStorage := js.Global().Get("localStorage")
	if !localStorage.IsUndefined() {
		localStorage.Call("setItem", "language", locale)
	}

	js.Global().Get("document").Get("documentElement").Set("lang", locale)
	os.Setenv("LANG", locale)
}
