package utils

import (
	"flagged-it/internal/translations"
	"log"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/lang"
)

func LoadTranslation(locale string) {
	lIdx := slices.IndexFunc(translations.TranslationsInfo, func(t translations.TranslationInfo) bool {
		return t.Name == locale
	})
	if lIdx < 0 {
		lIdx = slices.IndexFunc(translations.TranslationsInfo, func(t translations.TranslationInfo) bool {
			return t.Name == "en"
		})
	}
	if lIdx >= 0 {
		tr := translations.TranslationsInfo[lIdx]
		content, err := translations.FS.ReadFile("translations/" + tr.TranslationFileName)
		if err == nil {
			name := lang.SystemLocale().LanguageString()
			lang.AddTranslations(fyne.NewStaticResource(name+".json", content))
			return
		}
		log.Printf("Error loading translation file %s: %s\n", tr.TranslationFileName, err.Error())
	}
	if err := lang.AddTranslationsFS(translations.FS, "translations"); err != nil {
		log.Printf("Error loading translations: %s", err.Error())
	}
}
