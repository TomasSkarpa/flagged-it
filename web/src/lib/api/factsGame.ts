import { getApiUrl, API_ENDPOINTS } from './config';

export interface GuessHistoryEntry {
	guess: string;
	fact: string;
	isCorrect?: boolean;
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
	isCorrect: boolean;
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

export async function startFactsGame(): Promise<FactsGameSession> {
	const response = await fetch(getApiUrl(API_ENDPOINTS.FACTS_START), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
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
	const response = await fetch(getApiUrl(API_ENDPOINTS.FACTS_GUESS), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({
			sessionId,
			countryName,
		}),
	});

	if (!response.ok) {
		throw new Error(`Failed to submit guess: ${response.statusText}`);
	}

	return response.json();
}

export async function nextRound(sessionId: string): Promise<FactsGameSession> {
	const response = await fetch(getApiUrl(API_ENDPOINTS.FACTS_NEXT), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({
			sessionId,
		}),
	});

	if (!response.ok) {
		throw new Error(`Failed to start next round: ${response.statusText}`);
	}

	return response.json();
}
