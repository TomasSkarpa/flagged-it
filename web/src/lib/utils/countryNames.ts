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
