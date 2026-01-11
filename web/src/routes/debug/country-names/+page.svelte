<script lang="ts">
	import { onMount } from 'svelte';
	import { getAllCountries } from '$lib/api/debug';
	import { locale } from '$lib/stores/locale';
	import type { Country } from '$lib/types';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import { getCountryNameForLocale } from '$lib/utils/countryNames';

	const languages = locale.getSupportedLocales();
	
	// Find initial language index based on stored locale
	const storedLocale = $locale;
	const initialIndex = languages.findIndex(l => l.code === storedLocale);
	
	let currentLanguageIndex = initialIndex >= 0 ? initialIndex : 0;
	let countries: Country[] = [];
	let isLoading = false; // Start as false, will be set to true when loading starts
	let error: string | null = null;

	$: currentLanguage = languages[currentLanguageIndex];
	$: currentLocale = currentLanguage?.code || 'en';

	onMount(async () => {
		await loadCountries();
	});

	async function loadCountries() {
		if (isLoading) {
			return; // Prevent concurrent loads
		}
		isLoading = true;
		error = null;
		try {
			// Fetch countries with the selected locale
			const result = await getAllCountries(currentLocale);
			
			if (!result || !result.countries || result.countries.length === 0) {
				countries = [];
				error = 'No countries found';
				return;
			}
			
			countries = result.countries.sort((a, b) => {
				const nameA = getCountryNameForLocale(a);
				const nameB = getCountryNameForLocale(b);
				return nameA.localeCompare(nameB, currentLocale);
			});
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load countries';
			console.error('Error loading countries:', err);
			countries = [];
		} finally {
			isLoading = false;
		}
	}

	async function nextLanguage() {
		if (currentLanguageIndex < languages.length - 1) {
			currentLanguageIndex++;
			await loadCountries();
		}
	}

	async function prevLanguage() {
		if (currentLanguageIndex > 0) {
			currentLanguageIndex--;
			await loadCountries();
		}
	}

	async function goToLanguage(index: number) {
		if (index >= 0 && index < languages.length) {
			currentLanguageIndex = index;
			await loadCountries();
		}
	}
</script>

<svelte:head>
	<title>Country Names Debug - Flagged It</title>
</svelte:head>

<div class="min-h-screen p-4 md:p-8">
	<div class="max-w-6xl mx-auto">
		<!-- Header -->
		<div class="mb-6">
			<h1 class="text-3xl font-bold text-sandy-light mb-2">
				<span class="emoji-blue mr-2">🌍</span> Country Names Debug
			</h1>
			<p class="text-text-muted">
				View all country names in different languages
			</p>
		</div>

		<!-- Language Navigation -->
		<div class="card-game mb-6">
			<div class="flex items-center justify-between mb-4">
				<div class="flex items-center gap-4">
					<button
						on:click={prevLanguage}
						disabled={currentLanguageIndex === 0}
						class="btn-icon p-2 disabled:opacity-30 disabled:cursor-not-allowed"
						aria-label="Previous language"
					>
						←
					</button>
					<div class="text-center">
						<h2 class="text-xl font-bold text-sandy-light">
							{currentLanguage?.name || 'Unknown'}
						</h2>
						<p class="text-sm text-text-muted">
							{currentLanguage?.code.toUpperCase() || ''} • {countries.length} countries
						</p>
					</div>
					<button
						on:click={nextLanguage}
						disabled={currentLanguageIndex === languages.length - 1}
						class="btn-icon p-2 disabled:opacity-30 disabled:cursor-not-allowed"
						aria-label="Next language"
					>
						→
					</button>
				</div>
				<div class="text-sm text-text-muted">
					{currentLanguageIndex + 1} / {languages.length}
				</div>
			</div>

			<!-- Language Quick Select -->
			<div class="flex flex-wrap gap-2">
				{#each languages as lang, index}
					<button
						on:click={() => goToLanguage(index)}
						class="px-3 py-1 rounded-lg border-2 transition-all text-sm font-medium
							{index === currentLanguageIndex 
								? 'bg-primary/20 border-primary text-primary' 
								: 'bg-transparent border-white/20 text-text-muted hover:border-white/40'}"
					>
						{lang.code.toUpperCase()}
					</button>
				{/each}
			</div>
		</div>

		<!-- Content -->
		{#if isLoading}
			<div class="flex flex-col items-center justify-center py-20">
				<LoadingSpinner />
				<p class="mt-4 text-text-muted">Loading countries...</p>
			</div>
		{:else if error}
			<div class="card-game bg-error/20 border-error">
				<p class="text-error font-semibold">{error}</p>
				<p class="text-sm text-text-muted mt-2">Check the browser console for more details.</p>
			</div>
		{:else if countries.length === 0}
			<div class="card-game">
				<p class="text-text-muted text-center">No countries found</p>
				<p class="text-sm text-text-muted text-center mt-2">Check the browser console for debugging information.</p>
			</div>
		{:else}
			<div class="card-game">
				<div class="mb-4">
					<h3 class="text-lg font-semibold text-sandy-light mb-2">
						All Countries ({countries.length})
					</h3>
					<p class="text-sm text-text-muted">
						Sorted alphabetically in {currentLanguage?.name || currentLocale}
					</p>
				</div>
				
				<div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3 max-h-[70vh] overflow-y-auto">
					{#each countries as country (country.cca2)}
						<div class="flex items-center gap-2 p-2 rounded-lg bg-white/5 border border-white/10 hover:bg-white/10 transition-colors">
							{#if country.cca2}
								<img 
									src="/assets/twemoji_flags_cca2/{country.cca2}.svg" 
									alt="{country.cca2} flag"
									class="w-6 h-4 object-cover rounded flex-shrink-0"
								/>
							{/if}
							<div class="flex-1 min-w-0">
								<p class="text-sm font-medium text-sandy-light truncate">
									{getCountryNameForLocale(country)}
								</p>
								<p class="text-xs text-text-muted truncate">
									{country.cca2} / {country.cca3}
								</p>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	:global(:root.light) .btn-icon {
		background-color: rgba(0, 0, 0, 0.08);
		border-color: rgba(0, 0, 0, 0.15);
		color: var(--color-text);
	}
	:global(:root.light) .btn-icon:hover:not(:disabled) {
		background-color: rgba(0, 0, 0, 0.12);
		border-color: var(--color-accent);
	}
</style>
