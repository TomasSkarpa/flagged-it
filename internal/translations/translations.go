package translations

import "embed"

//go:embed translations
var FS embed.FS

type TranslationInfo struct {
	Name                string
	DisplayName         string
	TranslationFileName string
}

var TranslationsInfo = []TranslationInfo{
	{Name: "en", DisplayName: "English", TranslationFileName: "en.json"},
	{Name: "es", DisplayName: "Español", TranslationFileName: "es.json"},
	{Name: "fr", DisplayName: "Français", TranslationFileName: "fr.json"},
	{Name: "de", DisplayName: "Deutsch", TranslationFileName: "de.json"},
	{Name: "nl", DisplayName: "Nederlands", TranslationFileName: "nl.json"},
	{Name: "nb", DisplayName: "Norsk Bokmål", TranslationFileName: "nb.json"},
	{Name: "da", DisplayName: "Dansk", TranslationFileName: "da.json"},
	{Name: "sv", DisplayName: "Svenska", TranslationFileName: "sv.json"},
	{Name: "fi", DisplayName: "Suomi", TranslationFileName: "fi.json"},
	{Name: "pt", DisplayName: "Português", TranslationFileName: "pt.json"},
	{Name: "tr", DisplayName: "Türkçe", TranslationFileName: "tr.json"},
	{Name: "ro", DisplayName: "Română", TranslationFileName: "ro.json"},
	{Name: "hu", DisplayName: "Magyar", TranslationFileName: "hu.json"},
	{Name: "hr", DisplayName: "Hrvatski", TranslationFileName: "hr.json"},
	{Name: "cs", DisplayName: "Čeština", TranslationFileName: "cs.json"},
	{Name: "sk", DisplayName: "Slovenčina", TranslationFileName: "sk.json"},
	{Name: "pl", DisplayName: "Polski", TranslationFileName: "pl.json"},
	{Name: "it", DisplayName: "Italiano", TranslationFileName: "it.json"},
	{Name: "id", DisplayName: "Bahasa Indonesia", TranslationFileName: "id.json"},
	{Name: "ms", DisplayName: "Bahasa Melayu", TranslationFileName: "ms.json"},
	{Name: "fil", DisplayName: "Filipino", TranslationFileName: "fil.json"},
	{Name: "sw", DisplayName: "Kiswahili", TranslationFileName: "sw.json"},
	{Name: "vi", DisplayName: "Tiếng Việt", TranslationFileName: "vi.json"},
}
