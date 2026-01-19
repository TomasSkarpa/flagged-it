import type { Country } from '$lib/types';
import { get } from 'svelte/store';
import { locale } from '$lib/stores/locale';

/**
 * Get country name - the backend already translates the common field based on the locale
 * sent in the API request, so we just return the common field directly.
 * 
 * @param country - Country object (common field is already translated by backend)
 * @returns Translated country name
 */
export function getCountryName(country: Country): string {
	if (!country.name) {
		return '';
	}
	// Backend already translates country.name.common based on locale sent in request
	return country.name.common || '';
}

/**
 * Get country name for display, automatically using the current locale from the store.
 * This is the recommended function to use in Svelte components.
 * 
 * @param country - Country object
 * @returns Translated country name (based on current locale from store)
 */
export function getCountryNameForLocale(country: Country): string {
	// Backend already translated based on locale, so we just return common
	// This function exists for API consistency and potential future client-side translation
	return getCountryName(country);
}

/**
 * Find a country by matching the input string against various country identifiers.
 * Matches against: common name, official name, CCA2, CCA3, and translated names.
 * 
 * @param countries - Array of countries to search
 * @param input - Input string to match (country name or code)
 * @param currentLocale - Optional locale for translated name matching
 * @returns Matching country or null if not found
 */
export function findCountryByName(
	countries: Country[],
	input: string,
	currentLocale?: string
): Country | null {
	if (!input || !countries || countries.length === 0) {
		return null;
	}

	const inputLower = input.trim().toLowerCase();
	if (inputLower === '') {
		return null;
	}

	// Get locale from store if not provided
	const localeToUse = currentLocale || get(locale) || 'en';

	return countries.find(c => {
		// Match common name
		if (c.name.common.toLowerCase() === inputLower) {
			return true;
		}

		// Match official name
		if (c.name.official.toLowerCase() === inputLower) {
			return true;
		}

		// Match CCA2 code (e.g., "US")
		if (c.cca2.toLowerCase() === inputLower) {
			return true;
		}

		// Match CCA3 code (e.g., "USA")
		if (c.cca3.toLowerCase() === inputLower) {
			return true;
		}

		// Match translated name (if available)
		if (c.name.translations && c.name.translations[localeToUse]) {
			const translatedName = c.name.translations[localeToUse].toLowerCase();
			if (translatedName === inputLower) {
				return true;
			}
		}

		// Match using getCountryNameForLocale (handles backend-translated names)
		const displayName = getCountryNameForLocale(c).toLowerCase();
		if (displayName === inputLower) {
			return true;
		}

		return false;
	}) || null;
}
