//go:build !js && !wasm
// +build !js,!wasm

package utils

import (
	"os"
	"strings"
)

func GetSystemLocale() string {
	if locale := os.Getenv("LANG"); locale != "" {
		return strings.Split(strings.Split(locale, ".")[0], "_")[0]
	}
	return "en"
}

func SetSystemLocale(locale string) {
	os.Setenv("LANG", locale)
}
