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
	import { activeDropdown, toggleDropdown, closeDropdown } from '$lib/stores/dropdown';

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
	let sortBy: 'alphabetical' | 'population' | 'area' = 'alphabetical';
	let sortOrder: 'asc' | 'desc' = 'asc';
	let sortDropdownRef: HTMLDivElement;
	let sortButtonRef: HTMLButtonElement;
	let dropdownStyle = '';
	
	// Combined sort value for display
	$: sortValue = `${sortBy}-${sortOrder}`;
	$: isSortDropdownOpen = $activeDropdown === 'sort';
	
	function toggleSortDropdown() {
		toggleDropdown('sort');
	}
	
	// Update dropdown position when it opens
	$: if (isSortDropdownOpen && sortButtonRef) {
		updateDropdownPosition();
	}
	
	function updateDropdownPosition() {
		if (!sortButtonRef) return;
		const rect = sortButtonRef.getBoundingClientRect();
		// Find the max-w-7xl container (the relative positioning parent)
		const container = sortButtonRef.closest('.max-w-7xl') as HTMLElement;
		if (!container) return;
		const containerRect = container.getBoundingClientRect();
		
		dropdownStyle = `top: ${rect.bottom - containerRect.top + 8}px; right: ${containerRect.right - rect.right}px;`;
	}
	
	function selectSort(value: string) {
		const [by, order] = value.split('-');
		if (by === 'alphabetical' || by === 'population' || by === 'area') {
			sortBy = by;
			sortOrder = order === 'desc' ? 'desc' : 'asc';
		}
		closeDropdown();
	}
	
	function handleSortClickOutside(event: MouseEvent) {
		if (sortDropdownRef && !sortDropdownRef.contains(event.target as Node) &&
		    sortButtonRef && !sortButtonRef.contains(event.target as Node)) {
			closeDropdown();
		}
	}

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
		let filtered = countries.filter(country => {
			const matchesSearch = !searchQuery || 
				country.name.common.toLowerCase().includes(searchQuery.toLowerCase()) ||
				country.name.official.toLowerCase().includes(searchQuery.toLowerCase()) ||
				country.capital?.some(c => c.toLowerCase().includes(searchQuery.toLowerCase()));
			
			// Match region by value (not translated label)
			const matchesRegion = !selectedRegion || country.region === selectedRegion;
			
			return matchesSearch && matchesRegion;
		});
		
		// Apply sorting
		filtered = [...filtered].sort((a, b) => {
			let comparison = 0;
			
			if (sortBy === 'alphabetical') {
				comparison = a.name.common.localeCompare(b.name.common);
			} else if (sortBy === 'population') {
				comparison = a.population - b.population;
			} else if (sortBy === 'area') {
				comparison = a.area - b.area;
			}
			
			return sortOrder === 'asc' ? comparison : -comparison;
		});
		
		filteredCountries = filtered;
	}
	

	function handleCountryClick(country: Country) {
		selectedCountry = country;
		showModal = true;
	}

	function closeModal() {
		showModal = false;
		selectedCountry = null;
	}
	
	function handleModalKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && showModal) {
			closeModal();
		}
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

	// Sort option translations
	$: sortOptions = {
		'alphabetical-asc': t('library.sort.alphabetical_asc', undefined, currentLocale),
		'alphabetical-desc': t('library.sort.alphabetical_desc', undefined, currentLocale),
		'population-asc': t('library.sort.population_asc', undefined, currentLocale),
		'population-desc': t('library.sort.population_desc', undefined, currentLocale),
		'area-asc': t('library.sort.area_asc', undefined, currentLocale),
		'area-desc': t('library.sort.area_desc', undefined, currentLocale)
	};
</script>

<svelte:head>
	<title>{libraryTitle} - Flagged It</title>
	<meta name="description" content={libraryDescription} />
</svelte:head>

<svelte:window on:click={handleSortClickOutside} on:keydown={handleModalKeydown} />

<div class="min-h-screen p-4 md:p-8 overflow-x-hidden">
	<div class="max-w-7xl mx-auto w-full relative">
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

				<!-- Region Filter - Full Width -->
				<div>
					<p class="block text-sm font-medium text-text-muted mb-2">
						{filterRegionLabel}
					</p>
					<RegionSelector
						regions={regions}
						bind:selected={selectedRegion}
					/>
				</div>

				<!-- Results Count and Controls -->
				<div class="flex items-center justify-between mt-6">
					<div class="text-sm text-text-muted">
						{resultsCountText}
					</div>
					
					<!-- Sort and View Toggle -->
					<div class="flex items-center gap-2">
						<!-- Sort Dropdown -->
						<div class="library-sort-selector">
							<button
								bind:this={sortButtonRef}
								on:click|stopPropagation={toggleSortDropdown}
								class="px-3 py-2 rounded-lg border-2 transition-all bg-white/5 border-white/20 text-text-muted hover:border-accent text-sm flex items-center gap-2"
								style="min-width: 200px;"
								aria-label="Select sort option"
								aria-expanded={isSortDropdownOpen}
							>
								<span class="flex-1 text-left">{sortOptions[sortValue]}</span>
								<svg 
									class="chevron transition-transform" 
									class:rotated={isSortDropdownOpen}
									xmlns="http://www.w3.org/2000/svg" 
									width="16" 
									height="16" 
									viewBox="0 0 24 24" 
									fill="none" 
									stroke="currentColor" 
									stroke-width="2" 
									stroke-linecap="round" 
									stroke-linejoin="round"
								>
									<polyline points="6 9 12 15 18 9"></polyline>
								</svg>
							</button>
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
				</div>
			</div>
		</div>

		<!-- Sort Dropdown (rendered outside card-game to avoid stacking context issues) -->
		{#if isSortDropdownOpen && sortButtonRef}
			<div class="sort-dropdown" bind:this={sortDropdownRef} style={dropdownStyle}>
				<button
					class="sort-dropdown-item"
					class:selected={sortValue === 'alphabetical-asc'}
					on:click|stopPropagation={() => selectSort('alphabetical-asc')}
				>
					<span>{sortOptions['alphabetical-asc']}</span>
					{#if sortValue === 'alphabetical-asc'}
						<svg class="check-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<polyline points="20 6 9 17 4 12"></polyline>
						</svg>
					{/if}
				</button>
				<button
					class="sort-dropdown-item"
					class:selected={sortValue === 'alphabetical-desc'}
					on:click|stopPropagation={() => selectSort('alphabetical-desc')}
				>
					<span>{sortOptions['alphabetical-desc']}</span>
					{#if sortValue === 'alphabetical-desc'}
						<svg class="check-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<polyline points="20 6 9 17 4 12"></polyline>
						</svg>
					{/if}
				</button>
				<button
					class="sort-dropdown-item"
					class:selected={sortValue === 'population-asc'}
					on:click|stopPropagation={() => selectSort('population-asc')}
				>
					<span>{sortOptions['population-asc']}</span>
					{#if sortValue === 'population-asc'}
						<svg class="check-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<polyline points="20 6 9 17 4 12"></polyline>
						</svg>
					{/if}
				</button>
				<button
					class="sort-dropdown-item"
					class:selected={sortValue === 'population-desc'}
					on:click|stopPropagation={() => selectSort('population-desc')}
				>
					<span>{sortOptions['population-desc']}</span>
					{#if sortValue === 'population-desc'}
						<svg class="check-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<polyline points="20 6 9 17 4 12"></polyline>
						</svg>
					{/if}
				</button>
				<button
					class="sort-dropdown-item"
					class:selected={sortValue === 'area-asc'}
					on:click|stopPropagation={() => selectSort('area-asc')}
				>
					<span>{sortOptions['area-asc']}</span>
					{#if sortValue === 'area-asc'}
						<svg class="check-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<polyline points="20 6 9 17 4 12"></polyline>
						</svg>
					{/if}
				</button>
				<button
					class="sort-dropdown-item"
					class:selected={sortValue === 'area-desc'}
					on:click|stopPropagation={() => selectSort('area-desc')}
				>
					<span>{sortOptions['area-desc']}</span>
					{#if sortValue === 'area-desc'}
						<svg class="check-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<polyline points="20 6 9 17 4 12"></polyline>
						</svg>
					{/if}
				</button>
			</div>
		{/if}

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
				{#each filteredCountries as country (country.cca2)}
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
			<div class="space-y-3 library-list">
				{#each filteredCountries as country (country.cca2)}
					<button
						on:click={() => handleCountryClick(country)}
						class="w-full text-left library-list-item"
					>
						<div class="card-game hover:shadow-glow transition-all">
							<div class="flex items-center gap-4">
								<img
									src="/assets/twemoji_flags_cca2/{country.cca2}.svg"
									alt="{country.name.common} flag"
									class="w-24 h-auto object-contain rounded flex-shrink-0"
									style="aspect-ratio: 4/3; max-height: 80px;"
									loading="lazy"
									decoding="async"
								/>
								<div class="flex-1">
									<h3 class="text-xl font-bold text-sandy-light mb-1 mt-0">
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
			<div class="card-variant-subtle">
				<div class="flex flex-col sm:flex-row gap-6">
					<div class="flex-shrink-0">
						<div class="rounded-lg p-4 flex items-center justify-center">
							<img
								src="/assets/twemoji_flags_cca2/{selectedCountry.cca2}.svg"
								alt="{selectedCountry.name.common} flag"
								class="w-32 h-auto rounded"
								loading="eager"
								decoding="async"
							/>
						</div>
					</div>
					<div class="flex-1">
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
			</div>

			<!-- Statistics -->
			<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
				<div>
					<p class="text-sm font-semibold text-text-muted uppercase tracking-wide mb-2">{populationLabel}</p>
					<div class="card-variant-subtle stat-value-container">
						<p class="text-2xl font-bold text-sandy-light">{formatNumber(selectedCountry.population)}</p>
					</div>
				</div>
				<div>
					<p class="text-sm font-semibold text-text-muted uppercase tracking-wide mb-2">{areaLabel}</p>
					<div class="card-variant-subtle stat-value-container">
						<p class="text-2xl font-bold text-sandy-light">{formatArea(selectedCountry.area)}</p>
					</div>
				</div>
			</div>

			<!-- Capital -->
			{#if selectedCountry.capital && selectedCountry.capital.length > 0}
				<div>
					<p class="text-sm font-semibold text-text-muted uppercase tracking-wide mb-2">{capitalLabel}</p>
					<div class="card-variant-subtle stat-value-container">
						<p class="text-lg text-sandy-light">{selectedCountry.capital.join(', ')}</p>
					</div>
				</div>
			{/if}

			<!-- Subregion -->
			{#if selectedCountry.subregion}
				<div>
					<p class="text-sm font-semibold text-text-muted uppercase tracking-wide mb-2">{subregionLabel}</p>
					<div class="card-variant-subtle stat-value-container">
						<p class="text-lg text-sandy-light">{selectedCountry.subregion}</p>
					</div>
				</div>
			{/if}

			<!-- Languages -->
			{#if selectedCountry.languages && Object.keys(selectedCountry.languages).length > 0}
				<div>
					<p class="text-sm font-semibold text-text-muted uppercase tracking-wide mb-2">{languagesLabel}</p>
					<div class="card-variant-subtle stat-value-container">
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
					<div class="card-variant-subtle stat-value-container">
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
		will-change: scroll-position;
		width: 100%;
		max-width: 100%;
		overflow-x: hidden;
		overflow-y: visible;
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
		min-width: 0;
		height: 100%;
		text-align: left;
		background: transparent;
		border: none;
		padding: 8px 4px 4px 4px;
		cursor: pointer;
		overflow: visible;
	}
	
	.library-grid-item:focus {
		outline: none;
	}
	
	.library-grid-item:focus :global(.card-country-library) {
		border-color: var(--color-accent);
		background-color: var(--color-surface-light);
		box-shadow: 0 8px 24px 0 rgba(6, 182, 212, 0.2);
	}
	
	:global(:root.light) .library-grid-item:focus :global(.card-country-library) {
		border-color: var(--color-accent) !important;
		box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12) !important;
	}
	
	.library-list {
		will-change: scroll-position;
	}
	
	/* Sort selector dropdown styling */
	.library-sort-selector {
		position: relative;
	}
	
	.chevron {
		transition: transform 0.2s;
		flex-shrink: 0;
	}
	
	.chevron.rotated {
		transform: rotate(180deg);
	}
	
	.sort-dropdown {
		position: absolute;
		min-width: 200px;
		background: var(--color-surface);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 0.75rem;
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.4);
		overflow: hidden;
		animation: dropdownIn 0.2s ease-out;
		z-index: 1000;
		padding: 0.25rem;
	}
	
	:global(:root.light) .sort-dropdown {
		border-color: rgba(0, 0, 0, 0.25);
		border-width: 1.5px;
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
		background: var(--color-surface);
	}
	
	@keyframes dropdownIn {
		from {
			opacity: 0;
			transform: translateY(-8px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
	
	.sort-dropdown-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		width: 100%;
		padding: 0.625rem 0.75rem;
		border: none;
		background: transparent;
		color: var(--color-text);
		font-size: 0.875rem;
		text-align: left;
		cursor: pointer;
		border-radius: 0.5rem;
		transition: all 0.15s;
	}
	
	:global(:root.light) .sort-dropdown-item {
		color: var(--color-text);
	}
	
	.sort-dropdown-item:hover {
		background: rgba(255, 255, 255, 0.08);
	}
	
	:global(:root.light) .sort-dropdown-item:hover {
		background: rgba(0, 0, 0, 0.12);
	}
	
	.sort-dropdown-item.selected {
		background: rgba(99, 102, 241, 0.15);
		color: var(--color-primary-light);
	}
	
	:global(:root.light) .sort-dropdown-item.selected {
		background: rgba(99, 102, 241, 0.2);
		color: var(--color-primary-dark);
	}
	
	.check-icon {
		color: var(--color-primary);
		flex-shrink: 0;
	}
	
	/* Statistics value containers - handle overflow gracefully */
	.stat-value-container {
		min-width: 0;
		overflow-wrap: break-word;
		word-break: break-word;
	}
	
	.stat-value-container p {
		overflow-wrap: break-word;
		word-break: break-word;
		hyphens: auto;
		line-height: 1.2;
	}
	
	/* On very small screens, reduce font size for very long numbers */
	@media (max-width: 640px) {
		.stat-value-container p {
			font-size: 1.25rem;
		}
	}
	
</style>
