package utils

import (
	"flagged-it/internal/data/models"
	"testing"
)

func netherlandsFixture() models.Country {
	return models.Country{
		Name: models.CountryName{
			Common:   "Netherlands",
			Official: "Kingdom of the Netherlands",
			Translations: map[string]string{
				"el": "Ολλανδία",
				"en": "Netherlands",
			},
		},
		CCA2: "NL",
		CCA3: "NLD",
	}
}

func TestFindCountryByGuess_GreekLocale(t *testing.T) {
	countries := []models.Country{netherlandsFixture()}
	tests := []struct {
		guess    string
		locale   string
		wantCCA2 string
	}{
		{"Ολλανδία", "el", "NL"},
		{"ολλανδία", "el", "NL"},
		{"Netherlands", "el", "NL"},
		{"NL", "el", "NL"},
		{"nld", "el", "NL"},
		{"Kingdom of the Netherlands", "el", "NL"},
	}
	for _, tt := range tests {
		got := FindCountryByGuess(tt.guess, tt.locale, countries)
		if got == nil || got.CCA2 != tt.wantCCA2 {
			var cca2 string
			if got != nil {
				cca2 = got.CCA2
			}
			t.Errorf("FindCountryByGuess(%q, %q) = %v; want CCA2 %q", tt.guess, tt.locale, cca2, tt.wantCCA2)
		}
	}
}

func TestFindCountryByGuess_EnglishFallbackWhenLocaleNonEnglish(t *testing.T) {
	c := netherlandsFixture()
	c.Name.Common = "ΤεστLand"
	c.Name.Translations["el"] = "Ολλανδία"
	countries := []models.Country{c}

	got := FindCountryByGuess("Netherlands", "el", countries)
	if got == nil || got.CCA2 != "NL" {
		t.Fatalf("expected en translation fallback to match Netherlands, got %#v", got)
	}
}

func TestFindCountryByGuess_EmptyLocaleSkipsTranslations(t *testing.T) {
	countries := []models.Country{netherlandsFixture()}
	if FindCountryByGuess("Ολλανδία", "", countries) != nil {
		t.Fatal("expected no match when locale empty and common is English")
	}
	if got := FindCountryByGuess("Netherlands", "", countries); got == nil || got.CCA2 != "NL" {
		t.Fatalf("expected English common to match, got %#v", got)
	}
}

func TestMatchesCountry_DelegatesToSameRules(t *testing.T) {
	c := netherlandsFixture()
	if !MatchesCountry("Ολλανδία", c, "el") {
		t.Fatal("MatchesCountry should accept Greek name with locale el")
	}
	if MatchesCountry("Ολλανδία", c, "") {
		t.Fatal("MatchesCountry with empty locale should not use translations map")
	}
}
