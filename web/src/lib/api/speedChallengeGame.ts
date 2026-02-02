import { getApiUrl, API_ENDPOINTS } from './config';
import { get } from 'svelte/store';
import { locale } from '../stores/locale';
import type { Country } from '../types';

// Helper to get current locale
function getCurrentLocale(): string {
	if (typeof window === 'undefined') return 'en';
	return get(locale) || 'en';
}

export interface SpeedChallengeSession {
	sessionId: string;
	currentRound: number;
	currentGameType: string;
	timeLimit: number;
	maxRounds: number;
	score: number;
	total: number;
}

export interface SpeedChallengeQuestion {
	gameType: string;
	questionId: string;
	// Flag game
	flagUrl?: string;
	options?: Country[];
	// Shape game
	geoJson?: any;
	// Capital game
	countryName?: string;
	countryCca2?: string;
	// Facts game
	fact?: string;
	// Higher/Lower game
	left?: any;
	right?: any;
	category?: string;
	valueLabel?: string;
}

export interface SpeedChallengeAnswer {
	correct: boolean;
	points: number;
	score: number;
	total: number;
	timeTaken: number;
	finished: boolean;
	currentRound: number;
	correctAnswer?: string;
	nextQuestion?: SpeedChallengeQuestion;
	nextGameType?: string;
	timeLimit?: number;
}

export interface SpeedChallengeScore {
	score: number;
	total: number;
	maxRounds: number;
	isComplete: boolean;
	roundHistory: Array<{
		roundNumber: number;
		gameType: string;
		correct: boolean;
		timeTaken: number;
		points: number;
	}>;
}

export async function startSpeedChallenge(
	timeLimit: number = 30,
	maxRounds: number = 10
): Promise<SpeedChallengeSession & { question: SpeedChallengeQuestion }> {
	const currentLocale = getCurrentLocale();
	const response = await fetch(getApiUrl(API_ENDPOINTS.SPEED_CHALLENGE_START), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({ timeLimit, maxRounds, locale: currentLocale }),
	});

	if (!response.ok) {
		const errorText = await response.text();
		throw new Error(`Failed to start game: ${response.statusText} - ${errorText}`);
	}

	const data = await response.json();
	return data;
}

export async function getQuestion(sessionId: string): Promise<SpeedChallengeQuestion & {
	currentRound: number;
	currentGameType: string;
	timeLimit: number;
	score: number;
	total: number;
}> {
	const currentLocale = getCurrentLocale();
	const url = `${getApiUrl(API_ENDPOINTS.SPEED_CHALLENGE_QUESTION)}?sessionId=${sessionId}&locale=${currentLocale}`;
	const response = await fetch(url);

	if (!response.ok) {
		throw new Error(`Failed to get question: ${response.statusText}`);
	}

	return response.json();
}

export async function submitAnswer(
	sessionId: string,
	questionId: string,
	answer: string | { higher?: boolean; lower?: boolean },
	timeTaken: number
): Promise<SpeedChallengeAnswer> {
	const currentLocale = getCurrentLocale();
	
	// Normalize answer format
	let answerPayload: string;
	if (typeof answer === 'string') {
		answerPayload = answer;
	} else if (answer.higher) {
		answerPayload = 'higher';
	} else if (answer.lower) {
		answerPayload = 'lower';
	} else {
		answerPayload = '';
	}

	const response = await fetch(getApiUrl(API_ENDPOINTS.SPEED_CHALLENGE_ANSWER), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({
			sessionId,
			questionId,
			answer: answerPayload,
			timeTaken,
			locale: currentLocale,
		}),
	});

	if (!response.ok) {
		throw new Error(`Failed to submit answer: ${response.statusText}`);
	}

	return response.json();
}

export async function getScore(sessionId: string): Promise<SpeedChallengeScore> {
	const url = `${getApiUrl(API_ENDPOINTS.SPEED_CHALLENGE_SCORE)}?sessionId=${sessionId}`;
	const response = await fetch(url);

	if (!response.ok) {
		throw new Error(`Failed to get score: ${response.statusText}`);
	}

	return response.json();
}
