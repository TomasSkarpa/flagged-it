package utils

import (
	"flagged-it/internal/data/models"
	"strings"
)

func MatchesCountry(guess string, country models.Country) bool {
	guess = strings.ToLower(strings.TrimSpace(guess))
	if guess == "" {
		return false
	}

	// Check common name
	if strings.EqualFold(guess, country.Name.Common) {
		return true
	}

	// Check official name
	if strings.EqualFold(guess, country.Name.Official) {
		return true
	}

	// Check country codes (CCA2 like US, and CCA3 like USA)
	if strings.EqualFold(guess, country.CCA2) {
		return true
	}
	if strings.EqualFold(guess, country.CCA3) {
		return true
	}

	return false
}

func MatchesCountryByName(guess string, countryName string) bool {
	return strings.EqualFold(strings.TrimSpace(guess), strings.TrimSpace(countryName))
}
