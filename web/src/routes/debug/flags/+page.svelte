<script lang="ts">
	import { onMount } from 'svelte';
	import { getAllCountries } from '$lib/api/debug';
	import type { Country } from '$lib/types';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';

	let countries: Country[] = [];
	let filteredCountries: Country[] = [];
	let currentIndex = 0;
	let isLoading = true;
	let error: string | null = null;
	let viewMode: 'single' | 'grid' = 'single';
	let searchQuery = '';
	let selectedRegion = '';

	$: currentCountry = filteredCountries[currentIndex];
	$: {
		filteredCountries = countries.filter(c => {
			const matchesSearch = searchQuery === '' || 
				c.name.common.toLowerCase().includes(searchQuery.toLowerCase()) ||
				c.cca2.toLowerCase().includes(searchQuery.toLowerCase()) ||
				c.cca3.toLowerCase().includes(searchQuery.toLowerCase());
			const matchesRegion = selectedRegion === '' || c.region === selectedRegion;
			return matchesSearch && matchesRegion;
		});
		// Reset to first item when filter changes
		if (currentIndex >= filteredCountries.length) {
			currentIndex = 0;
		}
	}

	onMount(async () => {
		try {
			const result = await getAllCountries();
			countries = result.countries;
			filteredCountries = countries;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load countries';
		} finally {
			isLoading = false;
		}
	});

	function next() {
		if (currentIndex < filteredCountries.length - 1) {
			currentIndex++;
		}
	}

	function prev() {
		if (currentIndex > 0) {
			currentIndex--;
		}
	}

	function goTo(index: number) {
		currentIndex = index;
		viewMode = 'single';
	}

	function handleKeydown(e: KeyboardEvent) {
		if (viewMode === 'single') {
			if (e.key === 'ArrowRight' || e.key === 'ArrowDown') next();
			if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') prev();
		}
	}

	const regions = ['Africa', 'Americas', 'Asia', 'Europe', 'Oceania'];
</script>

<svelte:head>
	<title>All Flags - Debug</title>
</svelte:head>

<svelte:window on:keydown={handleKeydown} />

<div class="min-h-screen p-4 md:p-8">
	<div class="max-w-6xl mx-auto">
		<!-- Header -->
		<div class="flex flex-col md:flex-row items-start md:items-center justify-between gap-4 mb-6">
			<div>
				<h1 class="text-3xl font-bold text-sandy-light mb-1">
					<span class="emoji-blue mr-2">🚩</span> All Flags
				</h1>
				<p class="text-text-muted">
					{filteredCountries.length} of {countries.length} countries
				</p>
			</div>
			<div class="flex gap-2">
				<button 
					class="px-4 py-2 rounded-lg font-semibold transition-all {viewMode === 'single' ? 'bg-accent text-white' : 'bg-white/10 text-sandy-light hover:bg-white/20'}"
					on:click={() => viewMode = 'single'}
				>
					Single
				</button>
				<button 
					class="px-4 py-2 rounded-lg font-semibold transition-all {viewMode === 'grid' ? 'bg-accent text-white' : 'bg-white/10 text-sandy-light hover:bg-white/20'}"
					on:click={() => viewMode = 'grid'}
				>
					Grid
				</button>
			</div>
		</div>

		<!-- Filters -->
		<div class="card-game mb-6">
			<div class="flex flex-col md:flex-row gap-4">
				<div class="flex-1">
					<label class="block text-sm font-semibold text-sandy-light mb-2">Search</label>
					<input 
						type="text" 
						class="input-base" 
						placeholder="Search by name or code..."
						bind:value={searchQuery}
					/>
				</div>
				<div class="md:w-48">
					<label class="block text-sm font-semibold text-sandy-light mb-2">Region</label>
					<select class="input-base" bind:value={selectedRegion}>
						<option value="">All Regions</option>
						{#each regions as region}
							<option value={region}>{region}</option>
						{/each}
					</select>
				</div>
			</div>
		</div>

		{#if isLoading}
			<div class="flex flex-col items-center justify-center py-20">
				<LoadingSpinner />
				<p class="mt-4 text-text-muted">Loading countries...</p>
			</div>
		{:else if error}
			<div class="p-4 bg-error/20 border border-error rounded-lg">
				<p class="text-error font-semibold">{error}</p>
			</div>
		{:else if viewMode === 'single' && currentCountry}
			<!-- Single View -->
			<div class="card-game">
				<div class="flex flex-col items-center">
					<div class="text-sm text-text-muted mb-4">
						{currentIndex + 1} / {filteredCountries.length}
					</div>
					
					<img 
						src="/assets/twemoji_flags_cca2/{currentCountry.cca2}.svg" 
						alt="{currentCountry.name.common} flag"
						class="w-80 h-auto mb-6"
					/>
					
					<h2 class="text-3xl font-bold text-sandy-light mb-2">{currentCountry.name.common}</h2>
					<p class="text-lg text-text-muted mb-1">{currentCountry.name.official}</p>
					<div class="flex gap-4 text-sm text-text-muted mb-6">
						<span class="bg-white/10 px-3 py-1 rounded-full">{currentCountry.cca2}</span>
						<span class="bg-white/10 px-3 py-1 rounded-full">{currentCountry.cca3}</span>
						<span class="bg-white/10 px-3 py-1 rounded-full">{currentCountry.region}</span>
					</div>

					<div class="flex gap-4">
						<button 
							class="btn-secondary px-8 py-3 text-lg disabled:opacity-30"
							on:click={prev}
							disabled={currentIndex === 0}
						>
							← Previous
						</button>
						<button 
							class="btn-primary px-8 py-3 text-lg disabled:opacity-30"
							on:click={next}
							disabled={currentIndex === filteredCountries.length - 1}
						>
							Next →
						</button>
					</div>
				</div>
			</div>
		{:else if viewMode === 'grid'}
			<!-- Grid View -->
			<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
				{#each filteredCountries as country, i}
					<button 
						class="card-game p-4 hover:ring-2 hover:ring-accent transition-all text-center"
						on:click={() => goTo(i)}
					>
						<img 
							src="/assets/twemoji_flags_cca2/{country.cca2}.svg" 
							alt="{country.name.common} flag"
							class="w-full h-auto mb-2"
						/>
						<p class="text-sm font-semibold text-sandy-light truncate">{country.name.common}</p>
						<p class="text-xs text-text-muted">{country.cca2}</p>
					</button>
				{/each}
			</div>
		{:else}
			<div class="text-center py-20 text-text-muted">
				No countries found matching your filters.
			</div>
		{/if}
	</div>
</div>
