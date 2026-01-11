package models

type CountryName struct {
	Common       string            `json:"common"`
	Official     string            `json:"official"`
	Translations map[string]string `json:"translations,omitempty"`
}

type Country struct {
	Name       CountryName       `json:"name"`
	CCA2       string            `json:"cca2"`
	CCA3       string            `json:"cca3"`
	Capital    []string          `json:"capital"`
	Region     string            `json:"region"`
	Subregion  string            `json:"subregion"`
	Languages  map[string]string `json:"languages"`
	Latlng     []float64         `json:"latlng"`
	Population int               `json:"population"`
	Area       float64           `json:"area"`
}

// GetTranslatedName returns the country name in the specified locale
// Falls back to English (common) name if translation is not available
func (c *Country) GetTranslatedName(locale string) string {
	if c.Name.Translations == nil {
		return c.Name.Common
	}

	// If translation exists for the locale, return it
	if translated, ok := c.Name.Translations[locale]; ok && translated != "" {
		return translated
	}

	// Fallback to English
	return c.Name.Common
}

type CountryFacts struct {
	Name  string   `json:"name"`
	Facts []string `json:"facts"`
}
