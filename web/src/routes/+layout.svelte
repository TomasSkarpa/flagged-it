<script lang="ts">
	// @ts-nocheck
	import '../app.css';
	import { locale } from '$lib/stores/locale';
	import { theme } from '$lib/stores/theme';
	import { getMetaDescription } from '$lib/translations/meta';
	import { Navigation } from '$lib/components/nav';
	import { onMount } from 'svelte';
	// svelte-ignore unused-export-let
	export let data: any = {};
	// svelte-ignore unused-export-let
	export let params: Record<string, string> = {};
	
	let currentLocale = 'en';
	
	// Subscribe to locale changes and update meta description
	onMount(() => {
		// Initialize theme on mount
		const storedTheme = localStorage.getItem('theme');
		if (storedTheme === 'light' || storedTheme === 'dark' || storedTheme === 'system') {
			theme.set(storedTheme);
		}
		
		// RTL languages
		const rtlLanguages = ['ar', 'he'];
		
		const unsubscribe = locale.subscribe(value => {
			currentLocale = value;
			updateMetaDescription();
			
			// Set text direction for RTL languages
			if (typeof document !== 'undefined') {
				if (rtlLanguages.includes(value)) {
					document.documentElement.dir = 'rtl';
				} else {
					document.documentElement.dir = 'ltr';
				}
			}
		});
		
		// Set initial direction
		if (typeof document !== 'undefined') {
			if (rtlLanguages.includes(currentLocale)) {
				document.documentElement.dir = 'rtl';
			} else {
				document.documentElement.dir = 'ltr';
			}
		}
		
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

