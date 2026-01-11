<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import CountryCard from '$lib/components/ui/CountryCard.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import RegionSelector from '$lib/components/ui/RegionSelector.svelte';
	import { getAllCountries } from '$lib/api/debug';
	import type { Country } from '$lib/types';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';

	let countries: Country[] = [];
	let filteredCountries: Country[] = [];
	let isLoading = true;
	let error: string | null = null;
	let searchQuery = '';
	let selectedRegion = '';
	let selectedCountry: Country | null = null;
	let showModal = false;
	let viewMode: 'grid' | 'list' = 'grid';
	let previousLocale: string | null = null;

	// Reactive translations
	$: currentLocale = $locale;
	$: libraryTitle = t('library.title', undefined, currentLocale);
	$: libraryDescription = t('library.description', undefined, currentLocale);
	$: searchLabel = t('library.search.label', undefined, currentLocale);
	$: searchPlaceholder = t('library.search.placeholder', undefined, currentLocale);
	$: filterRegionLabel = t('library.filter.region', undefined, currentLocale);
	$: loadingText = t('library.loading', undefined, currentLocale);
	$: emptyText = t('library.empty', undefined, currentLocale);
	$: peopleText = t('library.detail.people', undefined, currentLocale);
	
	// Modal detail labels - reactive to locale
	$: codeLabel = t('library.detail.code', undefined, currentLocale);
	$: regionLabel = t('library.detail.region', undefined, currentLocale);
	$: populationLabel = t('library.detail.population', undefined, currentLocale);
	$: areaLabel = t('library.detail.area', undefined, currentLocale);
	$: capitalLabel = t('library.detail.capital', undefined, currentLocale);
	$: subregionLabel = t('library.detail.subregion', undefined, currentLocale);
	$: languagesLabel = t('library.detail.languages', undefined, currentLocale);
	$: locationLabel = t('library.detail.location', undefined, currentLocale);
	
	// Results count text - reactive to locale
	$: resultsCountText = t('library.results', [filteredCountries.length, countries.length], currentLocale);
	
	// Region labels - reactive to locale
	$: regions = [
		{ value: 'Africa', label: t('region.africa', undefined, currentLocale) },
		{ value: 'Americas', label: t('region.americas', undefined, currentLocale) },
		{ value: 'Asia', label: t('region.asia', undefined, currentLocale) },
		{ value: 'Europe', label: t('region.europe', undefined, currentLocale) },
		{ value: 'Oceania', label: t('region.oceania', undefined, currentLocale) }
	];
	
	// Region translation map for displaying region names in list view
	$: regionTranslationMap = {
		'Africa': t('region.africa', undefined, currentLocale),
		'Americas': t('region.americas', undefined, currentLocale),
		'Asia': t('region.asia', undefined, currentLocale),
		'Europe': t('region.europe', undefined, currentLocale),
		'Oceania': t('region.oceania', undefined, currentLocale)
	};
	
	function getTranslatedRegion(region: string): string {
		return regionTranslationMap[region] || region;
	}

	let scrollPosition = 0;
	let bodyStyle: { overflow?: string; position?: string; top?: string; width?: string } = {};

	function lockScroll() {
		if (typeof document === 'undefined') return;
		
		const body = document.body;
		const html = document.documentElement;
		
		// Save current scroll position
		scrollPosition = window.pageYOffset || document.documentElement.scrollTop;
		
		// Lock scroll for all browsers
		bodyStyle = {
			overflow: body.style.overflow,
			position: body.style.position,
			top: body.style.top,
			width: body.style.width
		};
		
		// Apply scroll lock
		body.style.overflow = 'hidden';
		body.style.position = 'fixed';
		body.style.top = `-${scrollPosition}px`;
		body.style.width = '100%';
		
		// For iOS Safari - also lock html
		html.style.overflow = 'hidden';
		html.style.position = 'fixed';
		html.style.height = '100%';
	}

	function unlockScroll() {
		if (typeof document === 'undefined') return;
		
		const body = document.body;
		const html = document.documentElement;
		
		// Restore original styles
		body.style.overflow = bodyStyle.overflow || '';
		body.style.position = bodyStyle.position || '';
		body.style.top = bodyStyle.top || '';
		body.style.width = bodyStyle.width || '';
		
		// Restore html styles
		html.style.overflow = '';
		html.style.position = '';
		html.style.height = '';
		
		// Restore scroll position
		window.scrollTo(0, scrollPosition);
	}

	// Lock/unlock scroll when modal opens/closes
	$: {
		if (showModal) {
			lockScroll();
		} else {
			unlockScroll();
		}
	}

	onDestroy(() => {
		// Ensure scroll is unlocked when component is destroyed
		unlockScroll();
	});

	async function loadCountries() {
		try {
			isLoading = true;
			error = null;
			const response = await getAllCountries();
			countries = response.countries || [];
			filteredCountries = countries;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load countries';
			console.error('Error loading countries:', err);
		} finally {
			isLoading = false;
		}
	}

	onMount(async () => {
		previousLocale = currentLocale;
		await loadCountries();
	});

	// Reload countries when locale changes (but not on initial mount)
	$: if (currentLocale && previousLocale !== null && previousLocale !== currentLocale) {
		previousLocale = currentLocale;
		loadCountries();
	}

	$: {
		filteredCountries = countries.filter(country => {
			const matchesSearch = !searchQuery || 
				country.name.common.toLowerCase().includes(searchQuery.toLowerCase()) ||
				country.name.official.toLowerCase().includes(searchQuery.toLowerCase()) ||
				country.capital?.some(c => c.toLowerCase().includes(searchQuery.toLowerCase()));
			
			// Match region by value (not translated label)
			const matchesRegion = !selectedRegion || country.region === selectedRegion;
			
			return matchesSearch && matchesRegion;
		});
	}

	function handleCountryClick(country: Country) {
		selectedCountry = country;
		showModal = true;
	}

	function closeModal() {
		showModal = false;
		selectedCountry = null;
	}

	function formatNumber(num: number): string {
		return new Intl.NumberFormat().format(num);
	}

	function formatArea(area: number): string {
		if (area >= 1_000_000) {
			return `${(area / 1_000_000).toFixed(2)}M km²`;
		} else if (area >= 1_000) {
			return `${(area / 1_000).toFixed(1)}K km²`;
		}
		return `${formatNumber(area)} km²`;
	}
</script>

<svelte:head>
	<title>{libraryTitle} - Flagged It</title>
	<meta name="description" content={libraryDescription} />
</svelte:head>

<div class="min-h-screen p-4 md:p-8">
	<div class="max-w-7xl mx-auto">
		<!-- Header -->
		<div class="text-center mb-8">
			<h1 class="text-4xl md:text-5xl font-bold mb-4">
				<span class="mr-3">📚</span>
				<span class="gradient-text">{libraryTitle}</span>
			</h1>
			<p class="text-lg text-text-muted">
				{libraryDescription}
			</p>
		</div>

		<!-- Search and Filters -->
		<div class="card-game mb-6">
			<div class="space-y-4">
				<!-- Search Bar -->
				<div>
					<label for="search" class="block text-sm font-medium text-text-muted mb-2">
						{searchLabel}
					</label>
					<input
						id="search"
						type="text"
						placeholder={searchPlaceholder}
						bind:value={searchQuery}
						class="input-base w-full"
					/>
				</div>

				<!-- Region Filter and View Toggle -->
				<div class="flex flex-col sm:flex-row gap-4 items-start sm:items-end">
					<div class="flex-1">
						<label class="block text-sm font-medium text-text-muted mb-2">
							{filterRegionLabel}
						</label>
						<RegionSelector
							regions={regions}
							bind:selected={selectedRegion}
						/>
					</div>
					
					<!-- View Mode Toggle -->
					<div class="flex gap-2">
						<button
							on:click={() => viewMode = 'grid'}
							class="px-3 py-2 rounded-lg border-2 transition-all
								{viewMode === 'grid' 
									? 'bg-primary/20 border-primary text-primary' 
									: 'bg-white/5 border-white/20 text-text-muted hover:border-accent'}"
							aria-label={t('library.view.grid', undefined, currentLocale)}
							title={t('library.view.grid', undefined, currentLocale)}
						>
							<!-- Grid Icon -->
							<svg 
								xmlns="http://www.w3.org/2000/svg" 
								width="20" 
								height="20" 
								viewBox="0 0 24 24" 
								fill="none" 
								stroke="currentColor" 
								stroke-width="2" 
								stroke-linecap="round" 
								stroke-linejoin="round"
							>
								<rect x="3" y="3" width="7" height="7"></rect>
								<rect x="14" y="3" width="7" height="7"></rect>
								<rect x="14" y="14" width="7" height="7"></rect>
								<rect x="3" y="14" width="7" height="7"></rect>
							</svg>
						</button>
						<button
							on:click={() => viewMode = 'list'}
							class="px-3 py-2 rounded-lg border-2 transition-all
								{viewMode === 'list' 
									? 'bg-primary/20 border-primary text-primary' 
									: 'bg-white/5 border-white/20 text-text-muted hover:border-accent'}"
							aria-label={t('library.view.list', undefined, currentLocale)}
							title={t('library.view.list', undefined, currentLocale)}
						>
							<!-- List Icon -->
							<svg 
								xmlns="http://www.w3.org/2000/svg" 
								width="20" 
								height="20" 
								viewBox="0 0 24 24" 
								fill="none" 
								stroke="currentColor" 
								stroke-width="2" 
								stroke-linecap="round" 
								stroke-linejoin="round"
							>
								<line x1="8" y1="6" x2="21" y2="6"></line>
								<line x1="8" y1="12" x2="21" y2="12"></line>
								<line x1="8" y1="18" x2="21" y2="18"></line>
								<line x1="3" y1="6" x2="3.01" y2="6"></line>
								<line x1="3" y1="12" x2="3.01" y2="12"></line>
								<line x1="3" y1="18" x2="3.01" y2="18"></line>
							</svg>
						</button>
					</div>
				</div>

				<!-- Results Count -->
				<div class="text-sm text-text-muted">
					{resultsCountText}
				</div>
			</div>
		</div>

		<!-- Content -->
		{#if isLoading}
			<div class="flex flex-col items-center justify-center py-20">
				<LoadingSpinner />
				<p class="mt-4 text-text-muted">{loadingText}</p>
			</div>
		{:else if error}
			<div class="card-game text-center py-12">
				<p class="text-error text-lg">{error}</p>
			</div>
		{:else if filteredCountries.length === 0}
			<div class="card-game text-center py-12">
				<p class="text-text-muted text-lg">{emptyText}</p>
			</div>
		{:else if viewMode === 'grid'}
			<!-- Grid View -->
			<div class="library-grid">
				{#each filteredCountries as country}
					<button
						on:click={() => handleCountryClick(country)}
						class="library-grid-item"
					>
						<CountryCard
							countryName={country.name.common}
							flagUrl="/assets/twemoji_flags_cca2/{country.cca2}.svg"
							stat={`${formatNumber(country.population)} ${peopleText}`}
							selected={false}
						/>
					</button>
				{/each}
			</div>
		{:else}
			<!-- List View -->
			<div class="space-y-3">
				{#each filteredCountries as country}
					<button
						on:click={() => handleCountryClick(country)}
						class="w-full text-left"
					>
						<div class="card-game hover:shadow-glow transition-all">
							<div class="flex items-center gap-4">
								<img
									src="/assets/twemoji_flags_cca2/{country.cca2}.svg"
									alt="{country.name.common} flag"
									class="w-16 h-12 object-contain rounded"
								/>
								<div class="flex-1">
									<h3 class="text-xl font-bold text-sandy-light mb-1">
										{country.name.common}
									</h3>
									<p class="text-sm text-text-muted mb-2">
										{country.name.official}
									</p>
									<div class="flex flex-wrap gap-4 text-sm text-text-muted">
										{#if country.capital && country.capital.length > 0}
											<span>{capitalLabel}: {country.capital.join(', ')}</span>
										{/if}
										<span>{populationLabel}: {formatNumber(country.population)}</span>
										<span>{areaLabel}: {formatArea(country.area)}</span>
										<span>{getTranslatedRegion(country.region)}</span>
									</div>
								</div>
							</div>
						</div>
					</button>
				{/each}
			</div>
		{/if}
	</div>
</div>

<!-- Country Detail Modal -->
{#if selectedCountry}
	<Modal 
		title={selectedCountry.name.common}
		open={showModal}
		onClose={closeModal}
	>
		<div class="space-y-6">
			<!-- Flag and Basic Info -->
			<div class="flex flex-col sm:flex-row gap-6">
				<div class="flex-shrink-0">
					<div class="bg-white/10 rounded-lg p-4 flex items-center justify-center">
						<img
							src="/assets/twemoji_flags_cca2/{selectedCountry.cca2}.svg"
							alt="{selectedCountry.name.common} flag"
							class="w-32 h-auto rounded"
						/>
					</div>
				</div>
				<div class="flex-1">
					<h2 class="text-2xl font-bold text-sandy-light mb-2">
						{selectedCountry.name.common}
					</h2>
					<p class="text-text-muted mb-4">{selectedCountry.name.official}</p>
					<div class="grid grid-cols-2 gap-4">
						<div>
							<p class="text-xs text-text-muted uppercase tracking-wide mb-1">{codeLabel}</p>
							<p class="font-semibold text-sandy-light">{selectedCountry.cca2} / {selectedCountry.cca3}</p>
						</div>
						<div>
							<p class="text-xs text-text-muted uppercase tracking-wide mb-1">{regionLabel}</p>
							<p class="font-semibold text-sandy-light">{getTranslatedRegion(selectedCountry.region)}</p>
						</div>
					</div>
				</div>
			</div>

			<!-- Statistics -->
			<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
				<div class="card-variant-subtle">
					<p class="text-xs text-text-muted uppercase tracking-wide mb-2">{populationLabel}</p>
					<p class="text-2xl font-bold text-sandy-light">{formatNumber(selectedCountry.population)}</p>
				</div>
				<div class="card-variant-subtle">
					<p class="text-xs text-text-muted uppercase tracking-wide mb-2">{areaLabel}</p>
					<p class="text-2xl font-bold text-sandy-light">{formatArea(selectedCountry.area)}</p>
				</div>
			</div>

			<!-- Capital -->
			{#if selectedCountry.capital && selectedCountry.capital.length > 0}
				<div>
					<p class="text-sm font-semibold text-text-muted uppercase tracking-wide mb-2">{capitalLabel}</p>
					<div class="card-variant-subtle">
						<p class="text-lg text-sandy-light">{selectedCountry.capital.join(', ')}</p>
					</div>
				</div>
			{/if}

			<!-- Subregion -->
			{#if selectedCountry.subregion}
				<div>
					<p class="text-sm font-semibold text-text-muted uppercase tracking-wide mb-2">{subregionLabel}</p>
					<div class="card-variant-subtle">
						<p class="text-lg text-sandy-light">{selectedCountry.subregion}</p>
					</div>
				</div>
			{/if}

			<!-- Languages -->
			{#if selectedCountry.languages && Object.keys(selectedCountry.languages).length > 0}
				<div>
					<p class="text-sm font-semibold text-text-muted uppercase tracking-wide mb-2">{languagesLabel}</p>
					<div class="card-variant-subtle">
						<div class="flex flex-wrap gap-2">
							{#each Object.entries(selectedCountry.languages) as [code, name]}
								<span class="px-3 py-1 bg-primary/20 border border-primary/30 rounded-lg text-sm text-sandy-light">
									{name}
								</span>
							{/each}
						</div>
					</div>
				</div>
			{/if}

			<!-- Coordinates -->
			{#if selectedCountry.latlng && selectedCountry.latlng.length === 2}
				<div>
					<p class="text-sm font-semibold text-text-muted uppercase tracking-wide mb-2">{locationLabel}</p>
					<div class="card-variant-subtle">
						<p class="text-sandy-light">
							{selectedCountry.latlng[0].toFixed(4)}°N, {selectedCountry.latlng[1].toFixed(4)}°E
						</p>
					</div>
				</div>
			{/if}
		</div>
	</Modal>
{/if}

<style>
	.library-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1rem;
		align-items: stretch;
	}
	
	@media (min-width: 640px) {
		.library-grid {
			grid-template-columns: repeat(3, 1fr);
		}
	}
	
	@media (min-width: 768px) {
		.library-grid {
			grid-template-columns: repeat(4, 1fr);
		}
	}
	
	@media (min-width: 1024px) {
		.library-grid {
			grid-template-columns: repeat(5, 1fr);
		}
	}
	
	.library-grid-item {
		display: flex;
		flex-direction: column;
		width: 100%;
		height: 100%;
		text-align: left;
		background: transparent;
		border: none;
		padding: 0;
		cursor: pointer;
	}
	
	.library-grid-item:focus {
		outline: 2px solid var(--color-primary);
		outline-offset: 2px;
	}
</style>
