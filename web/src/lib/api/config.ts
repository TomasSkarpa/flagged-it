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
		
		// Production (Vercel): use your backend server
		// Update this to match your server's public IP or domain
		// For now using HTTP - switch to HTTPS after SSL setup
		const backendHost = '91.109.38.74'; // Your server's public IP or domain
		const useHttps = false; // Set to true after SSL certificate setup
		const protocol = useHttps ? 'https' : 'http';
		const backendPort = useHttps ? '443' : '8080'; // 443 for HTTPS via nginx, 8080 for HTTP
		
		// If using HTTPS on standard port, omit port number
		if (useHttps && backendPort === '443') {
			return `${protocol}://${backendHost}`;
		}
		
		return `${protocol}://${backendHost}:${backendPort}`;
	}
	
	// Fallback for SSR
	return 'http://localhost:8080';
}

export const API_BASE_URL = getApiBaseUrl();

export const API_ENDPOINTS = {
	// Flag game endpoints
	FLAG_START: '/api/game/flag/start',
	FLAG_QUESTION: '/api/game/flag/question',
	FLAG_ANSWER: '/api/game/flag/answer',
	FLAG_SCORE: '/api/game/flag/score',
	
	// Shape game endpoints
	SHAPE_START: '/api/game/shape/start',
	SHAPE_QUESTION: '/api/game/shape/question',
	SHAPE_ANSWER: '/api/game/shape/answer',
	SHAPE_SCORE: '/api/game/shape/score',
	
	// Capital game endpoints
	CAPITAL_START: '/api/game/capital/start',
	CAPITAL_QUESTION: '/api/game/capital/question',
	CAPITAL_ANSWER: '/api/game/capital/answer',
	CAPITAL_SCORE: '/api/game/capital/score',
	
	// Higher/Lower game endpoints
	HIGHER_LOWER_START: '/api/game/higherlower/start',
	HIGHER_LOWER_ANSWER: '/api/game/higherlower/answer',
	HIGHER_LOWER_SCORE: '/api/game/higherlower/score',
	
	// Worldle game endpoints
	WORLDLE_START: '/api/game/worldle/start',
	WORLDLE_GUESS: '/api/game/worldle/guess',
	WORLDLE_STATE: '/api/game/worldle/state',
	
	// Facts game endpoints
	FACTS_START: '/api/game/facts/start',
	FACTS_GUESS: '/api/game/facts/guess',
	FACTS_NEXT: '/api/game/facts/next',
	
	// Debug/Browse endpoints
	DEBUG_COUNTRIES: '/api/debug/countries',
	DEBUG_GEOJSON: '/api/debug/geojson',
	DEBUG_GEOJSON_ALL: '/api/debug/geojson/all',
	
	// Health check
	HEALTH: '/api/health',
} as const;

export function getApiUrl(endpoint: string): string {
	// Check environment variable first
	if (typeof import.meta !== 'undefined' && import.meta.env?.VITE_API_URL) {
		return `${import.meta.env.VITE_API_URL}${endpoint}`;
	}
	
	// Always compute dynamically to handle SSR and client-side correctly
	if (typeof window !== 'undefined') {
		const hostname = window.location.hostname;
		if (isLocalNetwork(hostname)) {
			return `http://${hostname}:8080${endpoint}`;
		}
		
		// Production: use your backend server
		const backendHost = '91.109.38.74'; // Your server's public IP or domain
		const useHttps = false; // Set to true after SSL certificate setup
		const protocol = useHttps ? 'https' : 'http';
		const backendPort = useHttps ? '443' : '8080'; // 443 for HTTPS via nginx, 8080 for HTTP
		
		if (useHttps && backendPort === '443') {
			return `${protocol}://${backendHost}${endpoint}`;
		}
		
		return `${protocol}://${backendHost}:${backendPort}${endpoint}`;
	}
	// Fallback for SSR
	return `http://localhost:8080${endpoint}`;
}
