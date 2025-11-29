//go:build !js || !wasm
// +build !js !wasm

package utils

func IsMobile() bool {
	return false
}
