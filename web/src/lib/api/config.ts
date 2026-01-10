// API Configuration
// Automatically detects environment and sets API base URL

function isLocalNetwork(hostname: string): boolean {
	// Check for localhost variants
	if (hostname === 'localhost' || hostname === '127.0.0.1') {
		return true;
	}
	
	// Check for local network IPs (192.168.x.x, 10.x.x.x, 172.16-31.x.x)
	if (/^192\.168\./.test(hostname)) return true;
	if (/^10\./.test(hostname)) return true;
	if (/^172\.(1[6-9]|2[0-9]|3[0-1])\./.test(hostname)) return true;
	
	return false;
}

function getApiBaseUrl(): string {
	// Check environment variable first (for Vercel or custom builds)
	if (typeof import.meta !== 'undefined' && import.meta.env?.VITE_API_URL) {
		return import.meta.env.VITE_API_URL;
	}
	
	// Check if we're in development (localhost or local network)
	if (typeof window !== 'undefined') {
		const hostname = window.location.hostname;
		
		// Development: use same hostname with port 8080 for API
		if (isLocalNetwork(hostname)) {
			return `http://${hostname}:8080`;
		}
		
		// Production (Vercel): use relative path which will be proxied by vercel.json
		// Vercel will proxy /api/* requests to your backend server at 91.109.38.74:8080
		// This avoids mixed content issues (HTTPS frontend -> HTTP backend)
		return '/api';
	}
	
	// Fallback for SSR
	return 'http://localhost:8080';
}

export const API_BASE_URL = getApiBaseUrl();

export const API_ENDPOINTS = {
	// Flag game endpoints (without /api prefix - getApiUrl adds it)
	FLAG_START: '/game/flag/start',
	FLAG_QUESTION: '/game/flag/question',
	FLAG_ANSWER: '/game/flag/answer',
	FLAG_SCORE: '/game/flag/score',
	
	// Shape game endpoints
	SHAPE_START: '/game/shape/start',
	SHAPE_QUESTION: '/game/shape/question',
	SHAPE_ANSWER: '/game/shape/answer',
	SHAPE_SCORE: '/game/shape/score',
	
	// Capital game endpoints
	CAPITAL_START: '/game/capital/start',
	CAPITAL_QUESTION: '/game/capital/question',
	CAPITAL_ANSWER: '/game/capital/answer',
	CAPITAL_SCORE: '/game/capital/score',
	
	// Higher/Lower game endpoints
	HIGHER_LOWER_START: '/game/higherlower/start',
	HIGHER_LOWER_ANSWER: '/game/higherlower/answer',
	HIGHER_LOWER_SCORE: '/game/higherlower/score',
	
	// Worldle game endpoints
	WORLDLE_START: '/game/worldle/start',
	WORLDLE_GUESS: '/game/worldle/guess',
	WORLDLE_STATE: '/game/worldle/state',
	
		// Facts game endpoints
		FACTS_START: '/game/facts/start',
		FACTS_GUESS: '/game/facts/guess',
		FACTS_SKIP: '/game/facts/skip',
		FACTS_NEXT: '/game/facts/next',
	
	// Debug/Browse endpoints
	DEBUG_COUNTRIES: '/debug/countries',
	DEBUG_GEOJSON: '/debug/geojson',
	DEBUG_GEOJSON_ALL: '/debug/geojson/all',
	
	// Health check
	HEALTH: '/health',
} as const;

export function getApiUrl(endpoint: string): string {
	// Check environment variable first
	if (typeof import.meta !== 'undefined' && import.meta.env?.VITE_API_URL) {
		const baseUrl = import.meta.env.VITE_API_URL;
		// If VITE_API_URL is a relative path (starts with /), use it directly with endpoint
		// If it's an absolute URL, append /api and endpoint
		if (baseUrl.startsWith('/')) {
			return `${baseUrl}${endpoint}`;
		} else {
			return `${baseUrl}/api${endpoint}`;
		}
	}
	
	// Always compute dynamically to handle SSR and client-side correctly
	if (typeof window !== 'undefined') {
		const hostname = window.location.hostname;
		
		if (isLocalNetwork(hostname)) {
			// Development: endpoints need /api prefix since we're calling localhost directly
			return `http://${hostname}:8080/api${endpoint}`;
		}
		
		// Production (Vercel): use api subdomain with SSL
		// api.flaggedit.app points to the backend server with HTTPS
		return `https://api.flaggedit.app/api${endpoint}`;
	}
	// Fallback for SSR
	return `http://localhost:8080/api${endpoint}`;
}
