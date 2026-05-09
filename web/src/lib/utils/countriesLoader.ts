import { getAllCountries } from '$lib/api/data';
import type { Country } from '$lib/types';

/**
 * Fetches the country catalog for API-backed games using the same locale as the UI.
 * Server responses bake translations into name.common and omit name.translations.
 */
export async function fetchCountriesListForLocale(locale: string): Promise<Country[]> {
	const result = await getAllCountries(locale);
	return result.countries;
}
