<script lang="ts">
	// @ts-nocheck
	import '../app.css';
	import { locale } from '$lib/stores/locale';
	import { theme } from '$lib/stores/theme';
	import { getMetaDescription } from '$lib/translations/meta';
	import { Navigation } from '$lib/components/nav';
	import { hasAppleEmoji } from '$lib/utils/platform';
	import { onMount } from 'svelte';
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
		
		const unsubscribe = locale.subscribe(value => {
			currentLocale = value;
			updateMetaDescription();
			
			// Set lang attribute and text direction for RTL languages
			if (typeof document !== 'undefined') {
				document.documentElement.lang = value;
				if (rtlLanguages.includes(value)) {
					document.documentElement.dir = 'rtl';
				} else {
					document.documentElement.dir = 'ltr';
				}
			}
		});
		
		updateMetaDescription();
		return unsubscribe;
	});
	
	function updateMetaDescription() {
		if (typeof document !== 'undefined') {
			const metaDesc = document.querySelector('meta[name="description"]');
			if (metaDesc) {
				metaDesc.setAttribute('content', getMetaDescription(currentLocale));
			}
		}
	}
</script>

<svelte:head>
	<meta name="description" content={getMetaDescription(currentLocale)} />
</svelte:head>

<Navigation />

<main>
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
</style>

