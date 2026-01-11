import { getApiUrl, API_ENDPOINTS } from './config';
import { get } from 'svelte/store';
import { locale } from '../stores/locale';

// Helper to get current locale
function getCurrentLocale(): string {
	if (typeof window === 'undefined') return 'en';
	return get(locale) || 'en';
}

export type HigherLowerCategory = 'population' | 'area' | 'continents';

export interface HigherLowerItem {
	name: string;
	value: number;
	cca2?: string;
	imageUrl?: string;
}

export interface HigherLowerComparison {
	left: HigherLowerItem;
	right: HigherLowerItem;
	category: HigherLowerCategory;
	valueLabel: string;
}

export interface HigherLowerStartResponse {
	sessionId: string;
	comparison: HigherLowerComparison;
	score: number;
	highScore: number;
}

export interface HigherLowerAnswerResponse {
	correct: boolean;
	leftValue: number;
	rightValue: number;
	score: number;
	highScore: number;
	gameOver: boolean;
	nextComparison?: HigherLowerComparison;
}

export async function startHigherLowerGame(category: HigherLowerCategory = 'population'): Promise<HigherLowerStartResponse> {
	const currentLocale = getCurrentLocale();
	const response = await fetch(getApiUrl(API_ENDPOINTS.HIGHER_LOWER_START), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ category, locale: currentLocale }),
	});
	if (!response.ok) {
		throw new Error('Failed to start game');
	}
	return response.json();
}

export async function submitHigherLowerAnswer(
	sessionId: string,
	answer: 'higher' | 'lower'
): Promise<HigherLowerAnswerResponse> {
	const currentLocale = getCurrentLocale();
	const response = await fetch(getApiUrl(API_ENDPOINTS.HIGHER_LOWER_ANSWER), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ sessionId, answer, locale: currentLocale }),
	});
	if (!response.ok) {
		throw new Error('Failed to submit answer');
	}
	return response.json();
}

export function formatValue(value: number, category: HigherLowerCategory): string {
	if (category === 'population') {
		if (value >= 1_000_000_000) {
			return (value / 1_000_000_000).toFixed(2) + 'B';
		} else if (value >= 1_000_000) {
			return (value / 1_000_000).toFixed(2) + 'M';
		} else if (value >= 1_000) {
			return (value / 1_000).toFixed(1) + 'K';
		}
		return value.toLocaleString();
	} else if (category === 'area') {
		if (value >= 1_000_000) {
			return (value / 1_000_000).toFixed(2) + 'M';
		} else if (value >= 1_000) {
			return (value / 1_000).toFixed(1) + 'K';
		}
		return value.toLocaleString();
	}
	return value.toString();
}

export function getCategoryLabel(category: HigherLowerCategory): string {
	switch (category) {
		case 'population':
			return 'Population';
		case 'area':
			return 'Area (km²)';
		case 'continents':
			return 'Countries';
		default:
			return 'Value';
	}
}

export function getCategoryDescription(category: HigherLowerCategory): string {
	switch (category) {
		case 'population':
			return 'Which country has more people?';
		case 'area':
			return 'Which country is larger?';
		case 'continents':
			return 'Which continent has more countries?';
		default:
			return 'Which is higher?';
	}
}
