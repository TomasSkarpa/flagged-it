// Shared TypeScript types
// This will be populated as you build the UI kit and games

export interface Country {
	name: {
		common: string;
		official: string;
		nativeName?: Record<string, { common: string; official: string }>;
		translations?: Record<string, string>; // Language code -> translated name
	};
	cca2: string;
	cca3: string;
	capital: string[];
	region: string;
	subregion: string;
	population: number;
	area: number;
	languages?: Record<string, string>;
	latlng?: number[];
}

// Export types that match your Go models for consistency
export type GameType = 'flag' | 'shape' | 'hangman' | 'facts' | 'higher_lower' | 'list' | 'guessing';


