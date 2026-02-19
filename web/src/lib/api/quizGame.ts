import { getApiUrl, API_ENDPOINTS } from './config';
import { get } from 'svelte/store';
import { locale } from '../stores/locale';

export type QuizGameType =
	| 'flag'
	| 'shape'
	| 'capital'
	| 'higher_lower'
	| 'hangman'
	| 'facts'
	| 'guessing';

export interface QuizStartResponse {
	quizSessionId: string;
	score: number;
	totalRounds: number;
	currentRound: number;
	gameType?: string;
	data?: unknown;
}

export interface QuizRoundResponse {
	score: number;
	totalRounds: number;
	currentRound: number;
	gameType?: string;
	data?: unknown;
	complete?: boolean;
}

export interface QuizSubmitResponse {
	correct: boolean;
	scoreDelta: number;
	score: number;
	totalRounds: number;
	currentRound: number;
	revealedAnswer?: unknown;
	complete?: boolean;
	nextGameType?: string;
	nextData?: unknown;
}

function getCurrentLocale(): string {
	if (typeof window === 'undefined') return 'en';
	return get(locale) || 'en';
}

/**
 * Starts a new quiz with the given round types.
 * @param roundTypes - Array of game type ids (e.g. ['flag', 'shape', 'facts'])
 * @param region - Optional region filter
 * @returns Quiz session and first round payload
 */
export async function startQuiz(
	roundTypes: QuizGameType[],
	region: string = ''
): Promise<QuizStartResponse> {
	const currentLocale = getCurrentLocale();
	const response = await fetch(getApiUrl(API_ENDPOINTS.QUIZ_START), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			roundTypes,
			locale: currentLocale,
			region: region || undefined
		})
	});

	if (!response.ok) {
		const text = await response.text();
		throw new Error(text || `Failed to start quiz: ${response.statusText}`);
	}

	return response.json();
}

/**
 * Fetches the current round payload for a quiz session.
 * @param quizSessionId - Session id from startQuiz
 * @returns Current round payload or complete state
 */
export async function getQuizRound(quizSessionId: string): Promise<QuizRoundResponse> {
	const response = await fetch(
		`${getApiUrl(API_ENDPOINTS.QUIZ_ROUND)}?quizSessionId=${encodeURIComponent(quizSessionId)}`
	);

	if (!response.ok) {
		throw new Error(`Failed to get round: ${response.statusText}`);
	}

	return response.json();
}

/**
 * Submits the answer for the current round.
 * @param quizSessionId - Session id
 * @param data - Game-specific submit payload (e.g. { cca2 } for flag, { guess } for facts)
 * @returns Result and optional next round payload
 */
export async function submitQuizRound(
	quizSessionId: string,
	data: Record<string, unknown>
): Promise<QuizSubmitResponse> {
	const currentLocale = getCurrentLocale();
	const response = await fetch(getApiUrl(API_ENDPOINTS.QUIZ_ROUND), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			quizSessionId,
			data,
			locale: currentLocale
		})
	});

	if (!response.ok) {
		const text = await response.text();
		throw new Error(text || `Failed to submit: ${response.statusText}`);
	}

	return response.json();
}
