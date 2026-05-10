import { getApiUrl, API_ENDPOINTS } from './config';
import { get } from 'svelte/store';
import { locale } from '../stores/locale';

export interface FlagColorQuestion {
	questionId: string;
	flagUrl: string;
	cca2: string;
	guessableId: string;
	countryName: string;
	difficulty: string;
	maxPointsRound: number;
}

export interface FlagColorStartSession {
	sessionId: string;
	question: FlagColorQuestion;
}

export interface FlagColorAnswerResult {
	pointsEarned: number;
	deltaE: number;
	correctHex: string;
	guessHex: string;
	score: number;
	total: number;
	finished: boolean;
	maxPointsPerRound: number;
}

function getCurrentLocale(): string {
	if (typeof window === 'undefined') return 'en';
	return get(locale) || 'en';
}

export async function startFlagColorGame(
	region: string = '',
	opts?: { locale?: string; roundCount?: number; difficulty?: 'easy' | 'hard' }
): Promise<FlagColorStartSession> {
	const currentLocale = opts?.locale ?? getCurrentLocale();
	const body: Record<string, unknown> = { region, locale: currentLocale };
	if (opts?.roundCount != null && opts.roundCount > 0) {
		body.roundCount = opts.roundCount;
	}
	if (opts?.difficulty) {
		body.difficulty = opts.difficulty;
	}
	const response = await fetch(getApiUrl(API_ENDPOINTS.FLAG_COLOR_START), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body),
	});

	if (!response.ok) {
		throw new Error(`Failed to start game: ${response.statusText}`);
	}

	return response.json();
}

export async function getFlagColorQuestion(sessionId: string): Promise<FlagColorQuestion> {
	const currentLocale = getCurrentLocale();
	const url = `${getApiUrl(API_ENDPOINTS.FLAG_COLOR_QUESTION)}?sessionId=${encodeURIComponent(sessionId)}&locale=${encodeURIComponent(currentLocale)}`;
	const response = await fetch(url);
	if (!response.ok) {
		throw new Error(`Failed to get question: ${response.statusText}`);
	}
	return response.json();
}

export async function submitFlagColorAnswer(
	sessionId: string,
	questionId: string,
	r: number,
	g: number,
	b: number
): Promise<FlagColorAnswerResult> {
	const response = await fetch(getApiUrl(API_ENDPOINTS.FLAG_COLOR_ANSWER), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ sessionId, questionId, r, g, b }),
	});

	if (!response.ok) {
		throw new Error(`Failed to submit answer: ${response.statusText}`);
	}

	return response.json();
}
