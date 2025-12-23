//go:build !js || !wasm
// +build !js !wasm

package utils

// GetSavedTheme returns "system" for native builds
func GetSavedTheme() string {
	return "system"
}

// SetSavedTheme is a no-op for native builds
func SetSavedTheme(theme string) {
	// No-op for native
}

// GetSystemTheme returns "dark" for native builds
func GetSystemTheme() string {
	return "dark"
}

// GetEffectiveTheme returns "dark" for native builds
func GetEffectiveTheme() string {
	return "dark"
}

// ReloadPage is a no-op for native builds
func ReloadPage() {
	// No-op for native
}

