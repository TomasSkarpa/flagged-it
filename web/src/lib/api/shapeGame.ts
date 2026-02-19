import { getApiUrl, API_ENDPOINTS } from './config';
import type { Country } from '../types';
import { get } from 'svelte/store';
import { locale } from '../stores/locale';

// Helper to get current locale
function getCurrentLocale(): string {
	if (typeof window === 'undefined') return 'en';
	return get(locale) || 'en';
}

export interface GeoJSONGeometry {
	type: string;
	coordinates: number[][][] | number[][][][];
}

export interface GeoJSONFeature {
	type: string;
	id: string;
	properties: {
		name: string;
	};
	geometry: GeoJSONGeometry;
}

export interface GeoJSON {
	type: string;
	features: GeoJSONFeature[];
}

export interface ShapeQuestion {
	geoJson: GeoJSON;
	options: Country[];
	questionId: string;
}

export interface ShapeAnswer {
	correct: boolean;
	correctCca2: string;
	correctName: string;
	score: number;
	total: number;
	finished: boolean;
}

export interface ShapeGameSession {
	sessionId: string;
	question: ShapeQuestion;
}

export async function startShapeGame(
	region: string = '',
	opts?: { roundCount?: number }
): Promise<ShapeGameSession> {
	const currentLocale = getCurrentLocale();
	const body: { region: string; locale: string; roundCount?: number } = {
		region,
		locale: currentLocale,
	};
	if (opts?.roundCount != null && opts.roundCount > 0) {
		body.roundCount = opts.roundCount;
	}
	const response = await fetch(getApiUrl(API_ENDPOINTS.SHAPE_START), {
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

export async function getShapeQuestion(sessionId: string): Promise<ShapeQuestion> {
	const currentLocale = getCurrentLocale();
	const url = `${getApiUrl(API_ENDPOINTS.SHAPE_QUESTION)}?sessionId=${sessionId}&locale=${currentLocale}`;
	const response = await fetch(url);

	if (!response.ok) {
		throw new Error(`Failed to get question: ${response.statusText}`);
	}

	return response.json();
}

export async function submitShapeAnswer(
	sessionId: string,
	questionId: string,
	answerCca2: string
): Promise<ShapeAnswer> {
	const currentLocale = getCurrentLocale();
	const response = await fetch(getApiUrl(API_ENDPOINTS.SHAPE_ANSWER), {
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

export async function getShapeScore(sessionId: string): Promise<{ score: number; total: number }> {
	const url = `${getApiUrl(API_ENDPOINTS.SHAPE_SCORE)}?sessionId=${sessionId}`;
	const response = await fetch(url);

	if (!response.ok) {
		throw new Error(`Failed to get score: ${response.statusText}`);
	}

	return response.json();
}
