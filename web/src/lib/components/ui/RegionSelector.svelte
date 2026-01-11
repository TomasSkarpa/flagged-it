<script>
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	
	export let regions = []; // Array of { value: string, label: string } or string
	export let selected = '';
	export let onSelect = null;
	
	// Reactive translations
	$: currentLocale = $locale;
	$: worldLabel = t('region.world', undefined, currentLocale);
	
	function handleSelect(regionValue) {
		const value = typeof regionValue === 'object' && regionValue.value ? regionValue.value : regionValue;
		selected = value;
		if (onSelect) onSelect(value);
	}
	
	// Normalize regions to objects with value and label
	$: normalizedRegions = regions.map(r => {
		if (typeof r === 'object' && r.value) {
			return r; // Already normalized
		}
		// If it's a string, find the matching translation
		const regionMap = {
			'Africa': 'region.africa',
			'Americas': 'region.americas',
			'Asia': 'region.asia',
			'Europe': 'region.europe',
			'Oceania': 'region.oceania'
		};
		const key = regionMap[r] || r;
		return {
			value: r,
			label: typeof key === 'string' && key.startsWith('region.') ? t(key, undefined, currentLocale) : r
		};
	});
</script>

<div class="flex flex-wrap gap-3">
	<button
		on:click={() => handleSelect('')}
		class="region-filter-btn px-6 py-3 rounded-lg border-2 transition-all font-medium text-base {selected === '' ? 'bg-primary/20 border-primary text-primary' : 'bg-transparent border-white/20 text-sandy hover:border-white/40'}"
	>
		{worldLabel}
	</button>
	{#each normalizedRegions as region}
		<button
			on:click={() => handleSelect(region.value)}
			class="region-filter-btn px-6 py-3 rounded-lg border-2 transition-all font-medium text-base {selected === region.value ? 'bg-primary/20 border-primary text-primary' : 'bg-transparent border-white/20 text-sandy hover:border-white/40'}"
		>
			{region.label}
		</button>
	{/each}
</div>

<style>
	:global(:root.light) .region-filter-btn {
		color: var(--color-text) !important;
		border-color: rgba(0, 0, 0, 0.2) !important;
	}

	:global(:root.light) .region-filter-btn.text-sandy {
		color: var(--color-text) !important;
	}

	:global(:root.light) .region-filter-btn:hover {
		border-color: rgba(0, 0, 0, 0.3) !important;
		background: rgba(0, 0, 0, 0.05) !important;
	}

	:global(:root.light) .region-filter-btn:hover.text-sandy {
		color: var(--color-text) !important;
	}

	/* Selected state - match view mode toggle style */
	:global(:root.light) .region-filter-btn.bg-primary\/20 {
		background-color: rgba(99, 102, 241, 0.15) !important; /* primary with opacity */
		color: var(--color-primary) !important;
		border-color: var(--color-primary) !important;
	}
</style>
