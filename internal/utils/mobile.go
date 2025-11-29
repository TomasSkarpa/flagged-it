//go:build js && wasm
// +build js,wasm

package utils

import (
	"syscall/js"
)

func IsMobile() bool {
	userAgentData := js.Global().Get("navigator").Get("userAgentData")
	if !userAgentData.IsUndefined() && !userAgentData.IsNull() {
		if mobile := userAgentData.Get("mobile"); !mobile.IsUndefined() {
			return mobile.Bool()
		}
	}
	return js.Global().Get("navigator").Get("maxTouchPoints").Int() > 0
}
