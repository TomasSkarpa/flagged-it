import { getApiUrl, API_ENDPOINTS } from './config';
import { get } from 'svelte/store';
import { locale } from '../stores/locale';

// Helper to get current locale
function getCurrentLocale(): string {
	if (typeof window === 'undefined') return 'en';
	return get(locale) || 'en';
}

export interface GuessHistoryEntry {
	guess: string;
	fact: string;
	isCorrect?: boolean;
	country?: {
		cca2: string;
		name: string;
		flagUrl: string;
	};
}

export interface FactsGameSession {
	sessionId: string;
	currentFact: string;
	factNumber: number;
	triesLeft: number;
	score: number;
	total: number;
	isComplete: boolean;
}

export interface GuessResponse {
	isValidGuess?: boolean;
	isCorrect: boolean;
	error?: string;
	triesLeft: number;
	score: number;
	total: number;
	isComplete: boolean;
	guessHistory: GuessHistoryEntry[];
	nextFact?: string;
	factNumber?: number;
	correctCountry?: {
		cca2: string;
		name: string;
		flagUrl: string;
	};
}

export async function startFactsGame(opts?: { roundCount?: number }): Promise<FactsGameSession> {
	const currentLocale = getCurrentLocale();
	const body: { locale: string; roundCount?: number } = { locale: currentLocale };
	if (opts?.roundCount != null && opts.roundCount > 0) {
		body.roundCount = opts.roundCount;
	}
	const response = await fetch(getApiUrl(API_ENDPOINTS.FACTS_START), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify(body),
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
	const response = await fetch(getApiUrl(API_ENDPOINTS.FACTS_GUESS), {
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

export async function skipRound(sessionId: string): Promise<GuessResponse> {
	const currentLocale = getCurrentLocale();
	const response = await fetch(getApiUrl(API_ENDPOINTS.FACTS_SKIP), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({
			sessionId,
			locale: currentLocale,
		}),
	});

	if (!response.ok) {
		throw new Error(`Failed to skip round: ${response.statusText}`);
	}

	return response.json();
}

export async function nextRound(sessionId: string): Promise<FactsGameSession> {
	const currentLocale = getCurrentLocale();
	const response = await fetch(getApiUrl(API_ENDPOINTS.FACTS_NEXT), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({
			sessionId,
			locale: currentLocale,
		}),
	});

	if (!response.ok) {
		throw new Error(`Failed to start next round: ${response.statusText}`);
	}

	return response.json();
}
