// API client for communicating with Go backend
// This will be used when you build the API layer

const API_BASE = import.meta.env.VITE_API_URL || '/api';

export async function apiRequest<T>(
	endpoint: string,
	options?: RequestInit
): Promise<T> {
	const response = await fetch(`${API_BASE}${endpoint}`, {
		headers: {
			'Content-Type': 'application/json',
			...options?.headers,
		},
		...options,
	});

	if (!response.ok) {
		throw new Error(`API request failed: ${response.statusText}`);
	}

	return response.json();
}

// Example API functions (to be implemented)
export const api = {
	countries: {
		list: () => apiRequest('/countries'),
		get: (id: string) => apiRequest(`/countries/${id}`),
	},
	games: {
		start: (type: string) => apiRequest(`/games/${type}/start`, { method: 'POST' }),
		guess: (gameId: string, guess: string) =>
			apiRequest(`/games/${gameId}/guess`, {
				method: 'POST',
				body: JSON.stringify({ guess }),
			}),
	},
};


