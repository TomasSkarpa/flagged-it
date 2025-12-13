package utils

var currentLocale string

func GetCurrentLocale() string {
	if currentLocale == "" {
		currentLocale = GetSystemLocale()
	}
	return currentLocale
}

func SetCurrentLocale(locale string) {
	currentLocale = locale
	SetSystemLocale(locale)
}
