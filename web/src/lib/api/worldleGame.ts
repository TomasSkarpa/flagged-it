import { getApiUrl, API_ENDPOINTS } from './config';
import { get } from 'svelte/store';
import { locale } from '../stores/locale';

// Helper to get current locale
function getCurrentLocale(): string {
	if (typeof window === 'undefined') return 'en';
	return get(locale) || 'en';
}

export interface GuessEntry {
	country: {
		cca2: string;
		name: string;
		flagUrl: string;
		continent: string;
		population: number;
		area: number;
	};
	isCorrect: boolean;
	continent: string;
	continentCorrect: boolean;
	population: {
		value: number;
		direction: 'higher' | 'lower' | 'correct';
		proximity: 'very_close' | 'close' | 'somewhat_close' | 'far' | 'correct';
	};
	area: {
		value: number;
		direction: 'higher' | 'lower' | 'correct';
		proximity: 'very_close' | 'close' | 'somewhat_close' | 'far' | 'correct';
	};
}

export interface WorldleGameSession {
	sessionId: string;
	guessCount: number;
	isComplete: boolean;
}

export interface GuessResponse {
	isValidGuess: boolean;
	error?: string;
	isCorrect: boolean;
	guessCount: number;
	isComplete: boolean;
	guessEntry?: GuessEntry;
	correctCountry?: {
		cca2: string;
		name: string;
		flagUrl: string;
	};
}

export interface GameState {
	sessionId: string;
	guesses: GuessEntry[];
	guessCount: number;
	isComplete: boolean;
}

export async function startWorldleGame(): Promise<WorldleGameSession> {
	const currentLocale = getCurrentLocale();
	const response = await fetch(getApiUrl(API_ENDPOINTS.WORLDLE_START), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({ locale: currentLocale }),
	});

	if (!response.ok) {
		throw new Error(`Failed to start game: ${response.statusText}`);
	}

	return response.json();
}

export async function submitGuess(
	sessionId: string,
	countryName: string
): Promise<GuessResponse> {
	const currentLocale = getCurrentLocale();
	const response = await fetch(getApiUrl(API_ENDPOINTS.WORLDLE_GUESS), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({
			sessionId,
			countryName,
			locale: currentLocale,
		}),
	});

	if (!response.ok) {
		throw new Error(`Failed to submit guess: ${response.statusText}`);
	}

	return response.json();
}

export async function getGameState(sessionId: string): Promise<GameState> {
	const currentLocale = getCurrentLocale();
	const url = `${getApiUrl(API_ENDPOINTS.WORLDLE_STATE)}?sessionId=${sessionId}&locale=${currentLocale}`;
	const response = await fetch(url);

	if (!response.ok) {
		throw new Error(`Failed to get game state: ${response.statusText}`);
	}

	return response.json();
}

export function formatNumber(num: number): string {
	if (num >= 1000000) {
		return (num / 1000000).toFixed(1) + 'M';
	} else if (num >= 1000) {
		return (num / 1000).toFixed(1) + 'K';
	}
	return num.toString();
}
