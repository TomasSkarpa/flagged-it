//go:build !js || !wasm
// +build !js !wasm

package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// getThemeConfigPath returns the path to the theme config file
func getThemeConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(configDir, "flagged-it")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(appDir, "theme.txt"), nil
}

// GetSavedTheme reads theme from config file
func GetSavedTheme() string {
	path, err := getThemeConfigPath()
	if err != nil {
		return "system"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "system"
	}

	theme := strings.TrimSpace(string(data))
	if theme == "" {
		return "system"
	}

	return theme
}

// SetSavedTheme saves theme to config file
func SetSavedTheme(theme string) {
	path, err := getThemeConfigPath()
	if err != nil {
		return
	}

	os.WriteFile(path, []byte(theme), 0644)
}

// GetSystemTheme returns "dark" for native builds (could be enhanced to detect OS theme)
func GetSystemTheme() string {
	// TODO: Could detect actual system theme on macOS/Windows/Linux
	return "dark"
}

// GetEffectiveTheme returns the actual theme to use
func GetEffectiveTheme() string {
	saved := GetSavedTheme()
	if saved == "system" {
		return GetSystemTheme()
	}
	return saved
}

// ReloadPage is a no-op for native builds
func ReloadPage() {
	// No-op for native - theme changes would need app restart
}

