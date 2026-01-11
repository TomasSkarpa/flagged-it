// Type definitions matching Go backend models

export interface CountryName {
	common: string;
	official: string;
	translations?: Record<string, string>; // Language code -> translated name
}

export interface Country {
	name: CountryName;
	cca2: string;
	cca3: string;
	capital: string[];
	region: string;
	subregion: string;
	languages: Record<string, string>;
	latlng: number[];
	population: number;
	area: number;
}
