package utils

import (
	"fyne.io/fyne/v2/lang"
)

// TranslateRegion translates a region name to the current locale
// Falls back to the original region name if translation is not available
func TranslateRegion(region string) string {
	// Map region names to translation keys
	regionKeyMap := map[string]string{
		"World":      "region.world",
		"Africa":     "region.africa",
		"Asia":       "region.asia",
		"Europe":     "region.europe",
		"Americas":   "region.americas",
		"Oceania":    "region.oceania",
		"Antarctica": "region.antarctica",
	}

	// Get translation key for this region
	key, exists := regionKeyMap[region]
	if !exists {
		// If no translation key exists, return the original region name
		return region
	}

	// Get translated region name with fallback to original
	translated := lang.X(key, region)
	return translated
}
