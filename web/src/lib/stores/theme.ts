import { writable } from 'svelte/store';

export type Theme = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'theme';

function getStoredTheme(): Theme {
	if (typeof window === 'undefined') return 'system';
	const stored = localStorage.getItem(STORAGE_KEY);
	if (stored === 'light' || stored === 'dark' || stored === 'system') {
		return stored;
	}
	return 'system';
}

function getSystemTheme(): 'light' | 'dark' {
	if (typeof window === 'undefined') return 'dark';
	return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function applyTheme(theme: Theme) {
	if (typeof document === 'undefined') return;
	
	const effectiveTheme = theme === 'system' ? getSystemTheme() : theme;
	document.documentElement.setAttribute('data-theme', effectiveTheme);
	document.documentElement.classList.remove('light', 'dark');
	document.documentElement.classList.add(effectiveTheme);
}

function createThemeStore() {
	const { subscribe, set } = writable<Theme>(getStoredTheme());

	// Apply initial theme
	if (typeof window !== 'undefined') {
		applyTheme(getStoredTheme());
		
		// Listen for system theme changes
		window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
			const currentTheme = localStorage.getItem(STORAGE_KEY) as Theme;
			if (currentTheme === 'system') {
				applyTheme('system');
			}
		});
	}

	return {
		subscribe,
		set: (theme: Theme) => {
			if (typeof window !== 'undefined') {
				localStorage.setItem(STORAGE_KEY, theme);
			}
			applyTheme(theme);
			set(theme);
		},
		toggle: () => {
			const current = getStoredTheme();
			const next: Theme = current === 'dark' ? 'light' : current === 'light' ? 'system' : 'dark';
			if (typeof window !== 'undefined') {
				localStorage.setItem(STORAGE_KEY, next);
			}
			applyTheme(next);
			set(next);
		}
	};
}

export const theme = createThemeStore();
