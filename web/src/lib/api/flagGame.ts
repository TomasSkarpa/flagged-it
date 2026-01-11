import { getApiUrl, API_ENDPOINTS } from './config';
import type { Country } from '../types';
import { get } from 'svelte/store';
import { locale } from '../stores/locale';

export interface GameQuestion {
	flagUrl: string;
	options: Country[];
	questionId: string;
}

export interface GameAnswer {
	correct: boolean;
	correctCca2: string;
	correctName: string;
	score: number;
	total: number;
	finished: boolean;
}

export interface GameSession {
	sessionId: string;
	question: GameQuestion;
}

export interface StartGameRequest {
	region?: string;
	locale?: string;
}

export interface SubmitAnswerRequest {
	sessionId: string;
	questionId: string;
	answerCca2: string;
	locale?: string;
}

// Helper to get current locale
function getCurrentLocale(): string {
	if (typeof window === 'undefined') return 'en';
	return get(locale) || 'en';
}

export async function startGame(region: string = ''): Promise<GameSession> {
	const currentLocale = getCurrentLocale();
	const response = await fetch(getApiUrl(API_ENDPOINTS.FLAG_START), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({ region, locale: currentLocale }),
	});

	if (!response.ok) {
		throw new Error(`Failed to start game: ${response.statusText}`);
	}

	return response.json();
}

export async function getQuestion(sessionId: string): Promise<GameQuestion> {
	const currentLocale = getCurrentLocale();
	const url = `${getApiUrl(API_ENDPOINTS.FLAG_QUESTION)}?sessionId=${sessionId}&locale=${currentLocale}`;
	const response = await fetch(url);

	if (!response.ok) {
		throw new Error(`Failed to get question: ${response.statusText}`);
	}

	return response.json();
}

export async function submitAnswer(
	sessionId: string,
	questionId: string,
	answerCca2: string
): Promise<GameAnswer> {
	const currentLocale = getCurrentLocale();
	const response = await fetch(getApiUrl(API_ENDPOINTS.FLAG_ANSWER), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({
			sessionId: sessionId,
			questionId: questionId,
			answerCca2: answerCca2,
			locale: currentLocale,
		}),
	});

	if (!response.ok) {
		throw new Error(`Failed to submit answer: ${response.statusText}`);
	}

	return response.json();
}

export async function getScore(sessionId: string): Promise<{ score: number; total: number }> {
	const url = `${getApiUrl(API_ENDPOINTS.FLAG_SCORE)}?sessionId=${sessionId}`;
	const response = await fetch(url);

	if (!response.ok) {
		throw new Error(`Failed to get score: ${response.statusText}`);
	}

	return response.json();
}
