<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import RegionDropdown from '$lib/components/ui/RegionDropdown.svelte';
	
	export let title: string;
	export let emoji: string = '';
	export let description: string;
	export let isLoading: boolean = false;
	export let error: string | null = null;
	export let showRegionSelector: boolean = true;
	export let regions: { value: string; label: string }[] = [];
	
	// Reactive translations - will update when locale changes
	$: currentLocale = $locale;
	
	// Default regions with translations
	$: defaultRegions = [
		{ value: '', label: t('region.world', undefined, currentLocale) },
		{ value: 'Africa', label: t('region.africa', undefined, currentLocale) },
		{ value: 'Americas', label: t('region.americas', undefined, currentLocale) },
		{ value: 'Asia', label: t('region.asia', undefined, currentLocale) },
		{ value: 'Europe', label: t('region.europe', undefined, currentLocale) },
		{ value: 'Oceania', label: t('region.oceania', undefined, currentLocale) }
	];
	
	$: finalRegions = regions.length > 0 ? regions : defaultRegions;
	export let selectedRegion: string = '';
	export let startButtonText: string = 'Start Game';
	export let loadingText: string = 'Starting...';
	export let showStartButton: boolean = true;
	
	// Use translations for default values if not provided
	$: finalStartButtonText = startButtonText === 'Start Game' ? t('game.setup.start_game', undefined, currentLocale) : startButtonText;
	$: finalLoadingText = loadingText === 'Starting...' ? t('game.setup.starting', undefined, currentLocale) : loadingText;
	$: selectRegionLabel = t('game.setup.select_region', undefined, currentLocale);

	const dispatch = createEventDispatcher<{
		start: { region?: string; category?: string; [key: string]: any };
		regionChange: { region: string };
	}>();

	export let customStartData: Record<string, any> = {};

	function handleStart() {
		const startData: any = {};
		if (showRegionSelector) {
			startData.region = selectedRegion;
		}
		// Merge any custom data passed as prop (make a copy to avoid mutation)
		if (customStartData && Object.keys(customStartData).length > 0) {
			Object.assign(startData, { ...customStartData });
		}
		dispatch('start', startData);
	}

	function handleRegionChange(region: string) {
		selectedRegion = region;
		dispatch('regionChange', { region: selectedRegion });
	}
</script>

<div class="text-center px-4 md:px-8">
	<h1 class="text-4xl md:text-6xl font-bold text-sandy-light mb-4">
		{#if emoji}
			<span class="mr-2 emoji-blue">{emoji}</span>
		{/if}
		<span class="gradient-text">{title}</span>
	</h1>
	<p class="text-lg text-text-muted mb-8">
		{description}
	</p>
	
	{#if showRegionSelector}
		<div class="card-game max-w-md mx-auto mb-8">
			<RegionDropdown
				regions={finalRegions}
				bind:selected={selectedRegion}
				label={selectRegionLabel}
				onSelect={handleRegionChange}
			/>
		</div>
	{/if}
	
	<!-- Custom content slot (for additional options) -->
	<slot name="options" />
	
	{#if showStartButton}
		<button 
			class="btn-primary px-12 py-4 text-xl font-bold"
			disabled={isLoading}
			on:click={handleStart}
		>
			{isLoading ? finalLoadingText : finalStartButtonText}
		</button>
	{/if}
	
	{#if error}
		<div class="mt-6 p-4 bg-error/20 border border-error rounded-lg max-w-md mx-auto">
			<p class="text-error font-semibold">{error}</p>
		</div>
	{/if}
</div>
