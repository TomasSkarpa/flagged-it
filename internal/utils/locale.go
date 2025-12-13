//go:build js && wasm
// +build js,wasm

package utils

import (
	"os"
	"strings"
	"syscall/js"
)

func GetSystemLocale() string {
	htmlLang := js.Global().Get("document").Get("documentElement").Get("lang")
	if !htmlLang.IsUndefined() && !htmlLang.IsNull() && htmlLang.String() != "" {
		return strings.Split(htmlLang.String(), "-")[0]
	}
	return "en"
}

func SetSystemLocale(locale string) {
	js.Global().Get("document").Get("documentElement").Set("lang", locale)
	os.Setenv("LANG", locale)
}
