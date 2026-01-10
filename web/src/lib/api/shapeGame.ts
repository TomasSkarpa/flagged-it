import { getApiUrl, API_ENDPOINTS } from './config';
import type { Country } from '../types';

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

export async function startShapeGame(region: string = ''): Promise<ShapeGameSession> {
	const response = await fetch(getApiUrl(API_ENDPOINTS.SHAPE_START), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({ region }),
	});

	if (!response.ok) {
		throw new Error(`Failed to start game: ${response.statusText}`);
	}

	return response.json();
}

export async function getShapeQuestion(sessionId: string): Promise<ShapeQuestion> {
	const url = `${getApiUrl(API_ENDPOINTS.SHAPE_QUESTION)}?sessionId=${sessionId}`;
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
	const response = await fetch(getApiUrl(API_ENDPOINTS.SHAPE_ANSWER), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({
			sessionId: sessionId,
			questionId: questionId,
			answerCca2: answerCca2,
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
