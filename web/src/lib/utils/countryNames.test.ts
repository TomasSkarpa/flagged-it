import { describe, expect, it } from 'vitest';
import type { Country } from '$lib/types';
import { findCountryByName } from './countryNames';

function country(partial: Pick<Country, 'name' | 'cca2' | 'cca3'>): Country {
	return {
		...partial,
		capital: [],
		region: '',
		subregion: '',
		languages: {},
		latlng: [],
		population: 0,
		area: 0
	};
}

describe('findCountryByName', () => {
	const countries: Country[] = [
		country({
			name: {
				common: 'Greece',
				official: 'Hellenic Republic',
				translations: { el: 'Ελλάδα', en: 'Greece' }
			},
			cca2: 'GR',
			cca3: 'GRC'
		})
	];

	it('returns null for empty input or list', () => {
		expect(findCountryByName(countries, '', 'en')).toBeNull();
		expect(findCountryByName([], 'Greece', 'en')).toBeNull();
	});

	it('matches common name', () => {
		expect(findCountryByName(countries, 'greece', 'en')?.cca2).toBe('GR');
	});

	it('matches official name', () => {
		expect(findCountryByName(countries, 'hellenic republic', 'en')?.cca2).toBe('GR');
	});

	it('matches translation for locale', () => {
		expect(findCountryByName(countries, 'Ελλάδα', 'el')?.cca2).toBe('GR');
	});

	it('matches cca2/cca3', () => {
		expect(findCountryByName(countries, 'gr', 'en')?.cca2).toBe('GR');
		expect(findCountryByName(countries, 'grc', 'en')?.cca2).toBe('GR');
	});
});
