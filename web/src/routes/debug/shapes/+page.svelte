<script lang="ts">
	import { onMount } from 'svelte';
	import { getAllCountriesWithGeo, getCountryGeoJSON } from '$lib/api/data';
	import type { Country } from '$lib/types';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import ShapeRenderer from '$lib/components/ShapeRenderer.svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';

	$: currentLocale = $locale;
	$: debugShapesTitle = t('debug.shapes.document_title', undefined, currentLocale);
	$: debugShapesDescription = t('debug.shapes.meta_description', undefined, currentLocale);

	let countries: Country[] = [];
	let filteredCountries: Country[] = [];
	let currentIndex = 0;
	let currentGeoJSON: any = null;
	let isLoading = true;
	let isLoadingShape = false;
	let error: string | null = null;
	let viewMode: 'single' | 'grid' = 'single';
	let searchQuery = '';
	let selectedRegion = '';
	
	// Cache for grid view shapes
	let shapeCache: Map<string, any> = new Map();

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
		if (currentIndex >= filteredCountries.length) {
			currentIndex = 0;
		}
	}

	onMount(async () => {
		try {
			const result = await getAllCountriesWithGeo();
			countries = result.countries;
			filteredCountries = countries;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load countries';
		} finally {
			isLoading = false;
		}
	});

	// Load shape when currentCountry changes
	$: if (currentCountry && !isLoading) {
		loadCurrentShape();
	}

	async function loadCurrentShape() {
		if (!currentCountry) return;
		isLoadingShape = true;
		try {
			// Check cache first
			if (shapeCache.has(currentCountry.cca3)) {
				currentGeoJSON = shapeCache.get(currentCountry.cca3);
			} else {
				currentGeoJSON = await getCountryGeoJSON(currentCountry.cca3);
				shapeCache.set(currentCountry.cca3, currentGeoJSON);
			}
		} catch (err) {
			console.error('Failed to load shape:', err);
			currentGeoJSON = null;
		} finally {
			isLoadingShape = false;
		}
	}

	// Load shape for grid thumbnail
	async function loadShapeForGrid(cca3: string): Promise<any> {
		if (shapeCache.has(cca3)) {
			return shapeCache.get(cca3);
		}
		try {
			const geoJSON = await getCountryGeoJSON(cca3);
			shapeCache.set(cca3, geoJSON);
			return geoJSON;
		} catch {
			return null;
		}
	}

	// Load all shapes for grid view when switching to grid
	async function loadGridShapes() {
		const promises = filteredCountries.map(c => loadShapeForGrid(c.cca3));
		await Promise.all(promises);
		// Trigger reactivity
		shapeCache = shapeCache;
	}

	// Load shapes when switching to grid view
	$: if (viewMode === 'grid' && !isLoading) {
		loadGridShapes();
	}

	async function next() {
		if (currentIndex < filteredCountries.length - 1) {
			currentIndex++;
			await loadCurrentShape();
		}
	}

	async function prev() {
		if (currentIndex > 0) {
			currentIndex--;
			await loadCurrentShape();
		}
	}

	async function goTo(index: number) {
		currentIndex = index;
		viewMode = 'single';
		await loadCurrentShape();
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
	<title>{debugShapesTitle}</title>
	<meta name="description" content={debugShapesDescription} />
</svelte:head>

<svelte:window on:keydown={handleKeydown} />

<div class="min-h-screen p-4 md:p-8">
	<div class="max-w-6xl mx-auto">
		<!-- Header -->
		<div class="flex flex-col md:flex-row items-start md:items-center justify-between gap-4 mb-6">
			<div>
				<h1 class="text-3xl font-bold text-sandy-light mb-1">
					<span class="emoji-blue mr-2">🗺️</span> All Shapes
				</h1>
				<p class="text-text-muted">
					{filteredCountries.length} of {countries.length} countries with GeoJSON
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
					<label for="search-shapes" class="block text-sm font-semibold text-sandy-light mb-2">Search</label>
					<input 
						id="search-shapes"
						type="text" 
						class="input-base" 
						placeholder="Search by name or code..."
						bind:value={searchQuery}
					/>
				</div>
				<div class="md:w-48">
					<label for="region-shapes" class="block text-sm font-semibold text-sandy-light mb-2">Region</label>
					<select id="region-shapes" class="input-base" bind:value={selectedRegion}>
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
					
					<div class="bg-white/5 rounded-lg p-6 mb-6">
						{#if isLoadingShape}
							<div class="w-80 h-60 flex items-center justify-center">
								<LoadingSpinner />
							</div>
						{:else if currentGeoJSON}
							<ShapeRenderer 
								geoJson={currentGeoJSON}
								width={320}
								height={240}
								fillColor="var(--color-sandy-light)"
							/>
						{:else}
							<div class="w-80 h-60 flex items-center justify-center text-text-muted">
								No shape data available
							</div>
						{/if}
					</div>
					
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
							disabled={currentIndex === 0 || isLoadingShape}
						>
							← Previous
						</button>
						<button 
							class="btn-primary px-8 py-3 text-lg disabled:opacity-30"
							on:click={next}
							disabled={currentIndex === filteredCountries.length - 1 || isLoadingShape}
						>
							Next →
						</button>
					</div>
				</div>
			</div>
		{:else if viewMode === 'grid'}
			<!-- Grid View -->
			<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
				{#each filteredCountries as country, i}
					<button 
						class="card-game p-3 hover:ring-2 hover:ring-accent transition-all text-center"
						on:click={() => goTo(i)}
					>
						<div class="bg-white/5 rounded-lg mb-2 aspect-square flex items-center justify-center p-4">
							{#if shapeCache.has(country.cca3)}
								<ShapeRenderer 
									geoJson={shapeCache.get(country.cca3)}
									width={140}
									height={140}
									fillColor="var(--color-sandy-light)"
								/>
							{:else}
								<div class="w-full h-full flex items-center justify-center">
									<LoadingSpinner />
								</div>
							{/if}
						</div>
						<p class="text-sm font-semibold text-sandy-light truncate">{country.name.common}</p>
						<p class="text-xs text-text-muted">{country.cca3}</p>
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
