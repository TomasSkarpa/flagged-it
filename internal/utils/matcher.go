package utils

import (
	"strings"

	"flagged-it/internal/data/models"
)

type MatchLevel int

const (
	MatchCommon MatchLevel = 1 << iota
	MatchOfficial
	MatchAbbreviation
	MatchAll = MatchCommon | MatchOfficial | MatchAbbreviation
)

func MatchCountry(input string, country models.Country, level MatchLevel) bool {
	input = strings.TrimSpace(strings.ToLower(input))

	if (level & MatchCommon) != 0 {
		if strings.EqualFold(input, country.Name.Common) {
			return true
		}
	}

	if (level & MatchOfficial) != 0 {
		if strings.EqualFold(input, country.Name.Official) {
			return true
		}
	}

	if (level & MatchAbbreviation) != 0 {
		if strings.EqualFold(input, country.CCA2) || strings.EqualFold(input, country.CCA3) {
			return true
		}
	}

	return false
}
