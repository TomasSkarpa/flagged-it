package components

import (
	"flagged-it/internal/translations"
	"flagged-it/internal/utils"
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// FlagFromCountryCode converts an ISO 3166-1 alpha-2 country code to a flag emoji
func FlagFromCountryCode(countryCode string) string {
	if len(countryCode) != 2 {
		return "🌐"
	}
	code := strings.ToUpper(countryCode)
	r1 := rune(code[0]) - 'A' + 0x1F1E6
	r2 := rune(code[1]) - 'A' + 0x1F1E6
	if r1 < 0x1F1E6 || r1 > 0x1F1FF || r2 < 0x1F1E6 || r2 > 0x1F1FF {
		return "🌐"
	}
	return string([]rune{r1, r2})
}

// LanguageToCountryCode maps language codes to country codes for flags
func LanguageToCountryCode(langCode string) string {
	mapping := map[string]string{
		"en": "GB", "es": "ES", "fr": "FR", "de": "DE", "nl": "NL",
		"nb": "NO", "da": "DK", "sv": "SE", "fi": "FI", "pt": "PT",
		"tr": "TR", "ro": "RO", "hu": "HU", "hr": "HR", "cs": "CZ",
		"sk": "SK", "pl": "PL", "it": "IT", "id": "ID", "ms": "MY",
		"fil": "PH", "sw": "TZ", "vi": "VN",
	}
	if code, ok := mapping[langCode]; ok {
		return code
	}
	return ""
}

// GetLanguageFlag returns the flag emoji for a language code
func GetLanguageFlag(langCode string) string {
	return FlagFromCountryCode(LanguageToCountryCode(langCode))
}

// GetLanguageButtonText returns the button text for current language (e.g., "🇬🇧 EN")
func GetLanguageButtonText(langCode string) string {
	flag := GetLanguageFlag(langCode)
	return fmt.Sprintf("%s %s", flag, strings.ToUpper(langCode))
}

// ShowLanguageSelector shows a clean paginated language selector
func ShowLanguageSelector(window fyne.Window, onLanguageChanged func()) {
	var popup *widget.PopUp

	allLanguages := translations.TranslationsInfo
	languagesPerPage := 6
	totalPages := (len(allLanguages) + languagesPerPage - 1) / languagesPerPage
	currentPage := 0
	selectedLangCode := ""

	// Find current language and its page
	currentLocale := utils.GetCurrentLocale()
	for i, tr := range allLanguages {
		if tr.Name == currentLocale {
			currentPage = i / languagesPerPage
			selectedLangCode = tr.Name
			break
		}
	}

	// Fixed dimensions
	const cardWidth = float32(360)
	const rowHeight = float32(48)

	// Language buttons container (will be rebuilt on page change)
	var langButtons []*widget.Button
	langButtonsContainer := container.NewVBox()

	// Page label
	pageLabel := widget.NewLabel("")
	pageLabel.Alignment = fyne.TextAlignCenter

	// Navigation buttons
	prevBtn := widget.NewButton("◀ Prev", nil)
	nextBtn := widget.NewButton("Next ▶", nil)

	// Action buttons
	selectBtn := widget.NewButton("  Confirm  ", nil)
	selectBtn.Importance = widget.HighImportance

	closeBtn := widget.NewButton("  Close  ", nil)

	// Update page function
	updatePage := func() {
		langButtonsContainer.RemoveAll()
		langButtons = nil

		startIdx := currentPage * languagesPerPage
		endIdx := startIdx + languagesPerPage
		if endIdx > len(allLanguages) {
			endIdx = len(allLanguages)
		}

		// Create language buttons for this page
		for _, tr := range allLanguages[startIdx:endIdx] {
			tr := tr // capture
			flag := GetLanguageFlag(tr.Name)
			btnText := fmt.Sprintf("%s  %s", flag, tr.DisplayName)

			btn := widget.NewButton(btnText, func() {
				selectedLangCode = tr.Name
				// Update all buttons to reflect selection
				for j, b := range langButtons {
					idx := startIdx + j
					if idx < len(allLanguages) {
						if allLanguages[idx].Name == selectedLangCode {
							b.Importance = widget.HighImportance
						} else {
							b.Importance = widget.MediumImportance
						}
						b.Refresh()
					}
				}
			})
			btn.Alignment = widget.ButtonAlignLeading // Align text to left

			// Highlight if selected
			if tr.Name == selectedLangCode {
				btn.Importance = widget.HighImportance
			} else {
				btn.Importance = widget.MediumImportance
			}

			langButtons = append(langButtons, btn)
			langButtonsContainer.Add(btn)
		}

		// Add spacer rows if needed to maintain consistent height
		for i := endIdx - startIdx; i < languagesPerPage; i++ {
			spacer := canvas.NewRectangle(color.Transparent)
			spacer.SetMinSize(fyne.NewSize(cardWidth-32, rowHeight))
			langButtonsContainer.Add(spacer)
		}

		// Update page label
		pageLabel.SetText(fmt.Sprintf("Page %d of %d", currentPage+1, totalPages))

		// Update nav button states
		if currentPage == 0 {
			prevBtn.Disable()
		} else {
			prevBtn.Enable()
		}
		if currentPage >= totalPages-1 {
			nextBtn.Disable()
		} else {
			nextBtn.Enable()
		}

		langButtonsContainer.Refresh()
	}

	// Wire up navigation
	prevBtn.OnTapped = func() {
		if currentPage > 0 {
			currentPage--
			updatePage()
		}
	}
	nextBtn.OnTapped = func() {
		if currentPage < totalPages-1 {
			currentPage++
			updatePage()
		}
	}

	// Wire up action buttons
	selectBtn.OnTapped = func() {
		if selectedLangCode != "" {
			utils.SetCurrentLocale(selectedLangCode)
			utils.LoadTranslation(selectedLangCode)
			popup.Hide()
			if onLanguageChanged != nil {
				onLanguageChanged()
			}
		}
	}

	closeBtn.OnTapped = func() {
		popup.Hide()
	}

	// Initialize first page
	updatePage()

	// === BUILD THE LAYOUT ===

	// Title
	title := widget.NewLabel("🌍 Select Language")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	// Separator line
	separator1 := canvas.NewRectangle(color.RGBA{100, 100, 100, 255})
	separator1.SetMinSize(fyne.NewSize(cardWidth, 1))

	separator2 := canvas.NewRectangle(color.RGBA{100, 100, 100, 255})
	separator2.SetMinSize(fyne.NewSize(cardWidth, 1))

	// Navigation row: [Prev] [Page X of Y] [Next]
	navRow := container.NewBorder(
		nil, nil,
		prevBtn,
		nextBtn,
		container.NewCenter(pageLabel),
	)

	// Action row: [Close] [Select]
	actionRow := container.NewGridWithColumns(2, closeBtn, selectBtn)

	// Build content
	content := container.NewVBox(
		title,
		separator1,
		langButtonsContainer,
		separator2,
		navRow,
		actionRow,
	)

	// Wrap in padding
	paddedContent := container.NewPadded(content)

	popup = widget.NewModalPopUp(paddedContent, window.Canvas())
	popup.Show()
}

// NewLanguageSelectorButton creates a button that opens the language selector
func NewLanguageSelectorButton(window fyne.Window, onLanguageChanged func()) *widget.Button {
	currentLocale := utils.GetCurrentLocale()
	buttonText := GetLanguageButtonText(currentLocale)

	btn := widget.NewButton(buttonText, func() {
		ShowLanguageSelector(window, onLanguageChanged)
	})

	return btn
}
