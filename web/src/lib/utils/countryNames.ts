import type { Country } from '$lib/types';
import { get } from 'svelte/store';
import { locale } from '$lib/stores/locale';
import { normalizeAnswerForCompare } from '$lib/utils/answerNormalize';

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

function countryMatchesNormalizedInput(
	normInput: string,
	c: Country,
	localeToUse: string
): boolean {
	if (normalizeAnswerForCompare(c.name.common) === normInput) {
		return true;
	}
	if (normalizeAnswerForCompare(c.name.official) === normInput) {
		return true;
	}
	const tr = c.name.translations;
	if (tr && localeToUse) {
		const primary = tr[localeToUse];
		if (primary && normalizeAnswerForCompare(primary) === normInput) {
			return true;
		}
		if (localeToUse !== 'en' && tr.en && normalizeAnswerForCompare(tr.en) === normInput) {
			return true;
		}
	}
	if (normalizeAnswerForCompare(c.cca2) === normInput) {
		return true;
	}
	if (normalizeAnswerForCompare(c.cca3) === normInput) {
		return true;
	}
	return false;
}

/**
 * Find a country by matching the input string against various country identifiers.
 * Rules align with server-side FindCountryByGuess (normalization, locale translation, en fallback).
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

	const normInput = normalizeAnswerForCompare(input);
	if (normInput === '') {
		return null;
	}

	const localeToUse = currentLocale || get(locale) || 'en';

	return countries.find((c) => countryMatchesNormalizedInput(normInput, c, localeToUse)) || null;
}
