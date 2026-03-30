<script lang="ts">
	// @ts-nocheck
	import '../app.css';
	import { locale } from '$lib/stores/locale';
	import { theme } from '$lib/stores/theme';
	import { getMetaDescription, getOGTitle, getOGDescription, getOGImage, getOGUrl } from '$lib/translations/meta';
	import { Navigation } from '$lib/components/nav';
	import { hasAppleEmoji } from '$lib/utils/platform';
	import { browser } from '$app/environment';
	import { onDestroy, onMount, tick } from 'svelte';
	import { page } from '$app/stores';

	$: is404Page = $page.status === 404;

	$: if (browser) {
		document.body.classList.toggle('page-404', is404Page);
	}

	onDestroy(() => {
		if (browser) document.body.classList.remove('page-404');
	});
	// These props are provided by SvelteKit but not used in this layout
	export const data: any = {};
	export const params: Record<string, string> = {};
	
	let currentLocale = 'en';
	
	function getStoredLocale(): string {
		if (typeof window === 'undefined') return 'en';
		return localStorage.getItem('locale') || (navigator.language || 'en').split('-')[0];
	}
	
	// Subscribe to locale changes and update meta description
	onMount(() => {
		// Initialize theme on mount
		const storedTheme = localStorage.getItem('theme');
		if (storedTheme === 'light' || storedTheme === 'dark' || storedTheme === 'system') {
			theme.set(storedTheme);
		}
		
		// Apply Twemoji font if device doesn't have Apple emoji
		if (typeof document !== 'undefined' && !hasAppleEmoji()) {
			document.documentElement.classList.add('use-twemoji');
		}
		
		// RTL languages
		const rtlLanguages = ['ar', 'he'];
		
		// Set initial locale and lang attribute
		if (typeof document !== 'undefined') {
			const initialLocale = getStoredLocale();
			currentLocale = initialLocale;
			document.documentElement.lang = initialLocale;
			if (rtlLanguages.includes(initialLocale)) {
				document.documentElement.dir = 'rtl';
			} else {
				document.documentElement.dir = 'ltr';
			}
		}
		
		const unsubscribe = locale.subscribe((value) => {
			currentLocale = value;

			// Set lang attribute and text direction for RTL languages
			if (typeof document !== 'undefined') {
				document.documentElement.lang = value;
				if (rtlLanguages.includes(value)) {
					document.documentElement.dir = 'rtl';
				} else {
					document.documentElement.dir = 'ltr';
				}
			}

			void tick().then(() => updateOGTags());
		});

		// Subscribe to page changes to update OG URL
		const unsubscribePage = page.subscribe((pageData) => {
			void tick().then(() => updateOGTags(pageData.url.pathname));
		});

		void tick().then(() => updateOGTags());
		
		return () => {
			unsubscribe();
			unsubscribePage();
		};
	});
	
	function updateOGTags(path?: string) {
		if (typeof document === 'undefined') return;

		// Get current path - use provided path or fallback to window.location
		const currentPath = path || (typeof window !== 'undefined' ? window.location.pathname : '/');
		const metaDescContent = document
			.querySelector('meta[name="description"]')
			?.getAttribute('content')
			?.trim();
		const ogDescription = metaDescContent || getOGDescription(currentLocale);
		const ogUrl = getOGUrl(currentPath);
		
		// Update OG title (use page title if available, otherwise default)
		const pageTitle = document.querySelector('title')?.textContent || '';
		const ogTitle = pageTitle ? getOGTitle(pageTitle.replace(' - Flagged It', '').replace(' - Country Guessing Games', ''), currentLocale) : getOGTitle(undefined, currentLocale);
		
		// Update OG tags
		const ogTitleEl = document.getElementById('og-title');
		if (ogTitleEl) ogTitleEl.setAttribute('content', ogTitle);
		
		const ogDescEl = document.getElementById('og-description');
		if (ogDescEl) ogDescEl.setAttribute('content', ogDescription);
		
		const ogUrlEl = document.getElementById('og-url');
		if (ogUrlEl) ogUrlEl.setAttribute('content', ogUrl);
		
		// Update Twitter tags
		const twitterTitleEl = document.getElementById('twitter-title');
		if (twitterTitleEl) twitterTitleEl.setAttribute('content', ogTitle);
		
		const twitterDescEl = document.getElementById('twitter-description');
		if (twitterDescEl) twitterDescEl.setAttribute('content', ogDescription);
	}
</script>

<Navigation />

<main class:main-404-flush={is404Page}>
	<slot />
</main>

<style>
	:global(body) {
		margin: 0;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
	}

	main {
		min-height: calc(100vh - 4rem);
	}

	main.main-404-flush {
		padding: 0;
		margin: 0;
		width: 100%;
		max-width: none;
	}

	:global(body.page-404) {
		padding: 0;
		margin: 0;
	}
</style>

