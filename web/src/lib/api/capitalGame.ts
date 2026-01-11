import { getApiUrl, API_ENDPOINTS } from './config';
import { get } from 'svelte/store';
import { locale } from '../stores/locale';

// Helper to get current locale
function getCurrentLocale(): string {
	if (typeof window === 'undefined') return 'en';
	return get(locale) || 'en';
}

export interface CapitalQuestion {
	countryName: string;
	countryCca2: string;
	options: string[];
	questionId: string;
}

export interface CapitalAnswer {
	correct: boolean;
	correctCapital: string;
	correctCountry: string;
	score: number;
	total: number;
}

export interface CapitalGameSession {
	sessionId: string;
	question: CapitalQuestion;
}

export async function startCapitalGame(region: string = ''): Promise<CapitalGameSession> {
	const currentLocale = getCurrentLocale();
	const response = await fetch(getApiUrl(API_ENDPOINTS.CAPITAL_START), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ region, locale: currentLocale }),
	});
	if (!response.ok) {
		throw new Error('Failed to start capital game');
	}
	return response.json();
}

export async function getCapitalQuestion(sessionId: string): Promise<CapitalQuestion> {
	const currentLocale = getCurrentLocale();
	const response = await fetch(`${getApiUrl(API_ENDPOINTS.CAPITAL_QUESTION)}?sessionId=${sessionId}&locale=${currentLocale}`);
	if (!response.ok) {
		throw new Error('Failed to get question');
	}
	return response.json();
}

export async function submitCapitalAnswer(
	sessionId: string,
	questionId: string,
	answer: string
): Promise<CapitalAnswer> {
	const currentLocale = getCurrentLocale();
	const response = await fetch(getApiUrl(API_ENDPOINTS.CAPITAL_ANSWER), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ sessionId, questionId, answer, locale: currentLocale }),
	});
	if (!response.ok) {
		throw new Error('Failed to submit answer');
	}
	return response.json();
}
