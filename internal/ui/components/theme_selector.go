package components

import (
	"flagged-it/internal/utils"

	"fyne.io/fyne/v2"
)

// NewThemeSelectorButton creates a button that cycles through theme options
func NewThemeSelectorButton(window fyne.Window, app fyne.App) *Button {
	currentTheme := utils.GetSavedTheme()
	btn := NewButton(getThemeButtonText(currentTheme), nil)

	btn.OnTapped = func() {
		// Cycle through themes: system -> dark -> light -> system
		themes := []string{"system", "dark", "light"}
		currentTheme := utils.GetSavedTheme()
		currentIndex := -1
		for i, t := range themes {
			if t == currentTheme {
				currentIndex = i
				break
			}
		}
		
		nextIndex := (currentIndex + 1) % len(themes)
		nextTheme := themes[nextIndex]
		
		// Save to localStorage
		utils.SetSavedTheme(nextTheme)
		
		// Update button text
		btn.SetText(getThemeButtonText(nextTheme))
		
		// Reload page to apply theme (simplest and most reliable)
		utils.ReloadPage()
	}

	return btn
}

func getThemeButtonText(theme string) string {
	switch theme {
	case "light":
		return "☀️ Light"
	case "dark":
		return "🌙 Dark"
	case "system":
		return "💻 System"
	default:
		return "💻 System"
	}
}

