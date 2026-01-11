<script lang="ts">
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	
	export let regions: { value: string; label: string }[] = [];
	export let selected: string = '';
	export let onSelect: ((region: string) => void) | null = null;
	export let label: string = '';
	
	let isOpen = false;
	let dropdownRef: HTMLDivElement;
	
	// Reactive translations
	$: currentLocale = $locale;
	$: defaultLabel = label || t('game.setup.select_region', undefined, currentLocale);
	
	// Get selected region label
	$: selectedRegionObj = regions.find(r => r.value === selected);
	$: selectedLabel = selectedRegionObj?.label || (regions[0]?.label || '');
	
	function toggleDropdown() {
		isOpen = !isOpen;
	}
	
	function selectRegion(regionValue: string) {
		selected = regionValue;
		if (onSelect) {
			onSelect(regionValue);
		}
		isOpen = false;
	}
	
	function handleClickOutside(event: MouseEvent) {
		if (dropdownRef && !dropdownRef.contains(event.target as Node)) {
			isOpen = false;
		}
	}
</script>

<svelte:window on:click={handleClickOutside} />

<div class="region-dropdown" bind:this={dropdownRef}>
	{#if label}
		<span id="region-label" class="block text-sm font-semibold text-sandy-light mb-2">
			{label}
		</span>
	{/if}
	
	<button 
		class="region-btn"
		on:click|stopPropagation={toggleDropdown}
		aria-labelledby={label ? "region-label" : undefined}
		aria-label={label ? undefined : "Select region"}
		aria-expanded={isOpen}
		type="button"
		id="region-selector"
	>
		<span class="region-label">{selectedLabel}</span>
		<svg 
			class="chevron" 
			class:rotated={isOpen}
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
	
	{#if isOpen}
		<div class="dropdown">
			<div class="dropdown-list">
				{#each regions as region}
					<button
						class="dropdown-item"
						class:selected={selected === region.value}
						on:click|stopPropagation={() => selectRegion(region.value)}
						type="button"
					>
						<span class="item-label">{region.label}</span>
						{#if selected === region.value}
							<svg class="check-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
								<polyline points="20 6 9 17 4 12"></polyline>
							</svg>
						{/if}
					</button>
				{/each}
			</div>
		</div>
	{/if}
</div>

<style>
	.region-dropdown {
		position: relative;
		width: 100%;
	}
	
	.region-btn {
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
		padding: 0.75rem 1rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
		background: rgba(255, 255, 255, 0.05);
		border-radius: 0.5rem;
		color: var(--color-text-light);
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
	}
	
	.region-btn:hover {
		background: rgba(255, 255, 255, 0.1);
		border-color: rgba(255, 255, 255, 0.2);
	}
	
	:global(:root.light) .region-btn {
		border-color: rgba(0, 0, 0, 0.15);
		background: rgba(0, 0, 0, 0.04);
	}
	
	:global(:root.light) .region-btn:hover {
		background: rgba(0, 0, 0, 0.08);
		border-color: rgba(0, 0, 0, 0.25);
	}
	
	.region-label {
		flex: 1;
		text-align: left;
	}
	
	.chevron {
		transition: transform 0.2s;
		flex-shrink: 0;
		margin-left: 0.5rem;
		color: var(--color-text-muted);
	}
	
	.chevron.rotated {
		transform: rotate(180deg);
	}
	
	.dropdown {
		position: absolute;
		top: calc(100% + 0.5rem);
		left: 0;
		right: 0;
		min-width: 100%;
		max-height: 320px;
		background: var(--color-surface);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 0.75rem;
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.4);
		overflow: hidden;
		animation: dropdownIn 0.2s ease-out;
		z-index: 100;
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
	
	.dropdown-list {
		max-height: 260px;
		overflow-y: auto;
		padding: 0.25rem;
	}
	
	.dropdown-list::-webkit-scrollbar {
		width: 6px;
	}
	
	.dropdown-list::-webkit-scrollbar-track {
		background: transparent;
	}
	
	.dropdown-list::-webkit-scrollbar-thumb {
		background: rgba(255, 255, 255, 0.2);
		border-radius: 3px;
	}
	
	.dropdown-item {
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
	
	.dropdown-item:hover {
		background: rgba(255, 255, 255, 0.08);
	}
	
	:global(:root.light) .dropdown-item:hover {
		background: rgba(0, 0, 0, 0.06);
	}
	
	.dropdown-item.selected {
		background: rgba(99, 102, 241, 0.15);
		color: var(--color-primary-light);
	}
	
	.item-label {
		flex: 1;
	}
	
	.check-icon {
		color: var(--color-primary);
		flex-shrink: 0;
	}
</style>
