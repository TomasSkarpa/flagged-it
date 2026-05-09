package utils

import (
	"flagged-it/internal/data/models"
)

// MATCHING: All server-side resolution of user-typed country names must go through
// FindCountryByGuess or MatchesCountry so normalization, locale translations, and en fallback stay consistent.

// FindCountryByGuess returns the first country in countries that matches the guess after normalization.
//
// Matching uses, in order: common name, official name, translations[locale] (if locale is non-empty),
// translations["en"] when locale is non-empty and not "en" (bilingual fallback), then CCA2 and CCA3.
// If locale is empty, translation map entries are not used (only common/official/codes from raw data).
func FindCountryByGuess(guess string, locale string, countries []models.Country) *models.Country {
	normGuess := NormalizeAnswerForCompare(guess)
	if normGuess == "" {
		return nil
	}
	for i := range countries {
		c := &countries[i]
		if countryMatchesNormalizedGuess(normGuess, c, locale) {
			return c
		}
	}
	return nil
}

func countryMatchesNormalizedGuess(normGuess string, country *models.Country, locale string) bool {
	if normGuess == NormalizeAnswerForCompare(country.Name.Common) {
		return true
	}
	if normGuess == NormalizeAnswerForCompare(country.Name.Official) {
		return true
	}
	if locale != "" && country.Name.Translations != nil {
		if translatedName, ok := country.Name.Translations[locale]; ok && translatedName != "" {
			if normGuess == NormalizeAnswerForCompare(translatedName) {
				return true
			}
		}
		if locale != "en" {
			if translatedName, ok := country.Name.Translations["en"]; ok && translatedName != "" {
				if normGuess == NormalizeAnswerForCompare(translatedName) {
					return true
				}
			}
		}
	}
	if normGuess == NormalizeAnswerForCompare(country.CCA2) {
		return true
	}
	if normGuess == NormalizeAnswerForCompare(country.CCA3) {
		return true
	}
	return false
}

// MatchesCountry reports whether guess refers to the given country using the same rules as FindCountryByGuess.
func MatchesCountry(guess string, country models.Country, locale string) bool {
	normGuess := NormalizeAnswerForCompare(guess)
	if normGuess == "" {
		return false
	}
	return countryMatchesNormalizedGuess(normGuess, &country, locale)
}

func MatchesCountryByName(guess string, countryName string) bool {
	return NormalizeAnswerForCompare(guess) == NormalizeAnswerForCompare(countryName)
}
