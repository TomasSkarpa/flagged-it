import type { Country } from '$lib/types';

/**
 * Get country name in the specified locale
 * Falls back to common name if translation is not available
 * 
 * @param country - Country object with translations
 * @param locale - Target locale code (e.g., 'en', 'es', 'fr')
 * @returns Translated country name or fallback to common name
 */
export function getCountryName(country: Country, locale: string): string {
	// Check if translations exist
	if (country.name && (country.name as any).translations) {
		const translations = (country.name as any).translations;
		// Return translation for the locale, or fallback to common name
		return translations[locale] || country.name.common;
	}
	// Fallback to common name
	return country.name.common;
}

/**
 * Get country name for display, using the current locale from the store
 * This is a reactive version that can be used in Svelte components
 * 
 * @param country - Country object
 * @param currentLocale - Current locale from locale store (should be reactive)
 * @returns Translated country name
 */
export function getCountryNameForLocale(country: Country, currentLocale: string): string {
	return getCountryName(country, currentLocale);
}
