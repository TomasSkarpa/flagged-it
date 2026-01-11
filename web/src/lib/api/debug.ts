import { getApiUrl, API_ENDPOINTS } from './config';
import type { Country } from '../types';
import { locale } from '../stores/locale';
import { get } from 'svelte/store';

export interface CountriesResponse {
	countries: Country[];
	total: number;
}

export async function getAllCountries(localeOverride?: string): Promise<CountriesResponse> {
	const currentLocale = localeOverride || get(locale);
	const url = `${getApiUrl(API_ENDPOINTS.DEBUG_COUNTRIES)}?locale=${currentLocale}`;
	const response = await fetch(url);
	if (!response.ok) {
		const errorText = await response.text();
		console.error('API Error:', errorText);
		throw new Error(`Failed to fetch countries: ${response.status} ${response.statusText}`);
	}
	const data = await response.json();
	return data;
}

export async function getCountryGeoJSON(cca3: string): Promise<any> {
	const response = await fetch(`${getApiUrl(API_ENDPOINTS.DEBUG_GEOJSON)}?cca3=${cca3}`);
	if (!response.ok) {
		throw new Error(`Failed to fetch GeoJSON for ${cca3}`);
	}
	return response.json();
}

export async function getAllCountriesWithGeo(): Promise<CountriesResponse> {
	const response = await fetch(getApiUrl(API_ENDPOINTS.DEBUG_GEOJSON_ALL));
	if (!response.ok) {
		throw new Error('Failed to fetch countries with GeoJSON');
	}
	return response.json();
}
