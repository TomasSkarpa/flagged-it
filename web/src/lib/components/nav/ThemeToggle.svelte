<script lang="ts">
	import { theme, type Theme } from '$lib/stores/theme';
	
	let isOpen = false;
	let dropdownRef: HTMLDivElement;
	
	const themes: { value: Theme; label: string; icon: string }[] = [
		{ value: 'light', label: 'Light', icon: '☀️' },
		{ value: 'dark', label: 'Dark', icon: '🌙' },
		{ value: 'system', label: 'System', icon: '💻' }
	];
	
	function toggleDropdown() {
		isOpen = !isOpen;
	}
	
	function selectTheme(value: Theme) {
		theme.set(value);
		isOpen = false;
	}
	
	function handleClickOutside(event: MouseEvent) {
		if (dropdownRef && !dropdownRef.contains(event.target as Node)) {
			isOpen = false;
		}
	}
	
	function getCurrentIcon() {
		const current = themes.find(t => t.value === $theme);
		return current?.icon || '🌙';
	}
</script>

<svelte:window on:click={handleClickOutside} />

<div class="theme-toggle" bind:this={dropdownRef}>
	<button 
		class="theme-btn"
		on:click|stopPropagation={toggleDropdown}
		aria-label="Select theme"
		aria-expanded={isOpen}
	>
		<span class="theme-icon">{getCurrentIcon()}</span>
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
			{#each themes as themeOption}
				<button
					class="dropdown-item"
					class:selected={$theme === themeOption.value}
					on:click|stopPropagation={() => selectTheme(themeOption.value)}
				>
					<span class="option-icon">{themeOption.icon}</span>
					<span class="option-label">{themeOption.label}</span>
					{#if $theme === themeOption.value}
						<svg class="check-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<polyline points="20 6 9 17 4 12"></polyline>
						</svg>
					{/if}
				</button>
			{/each}
		</div>
	{/if}
</div>

<style>
	.theme-toggle {
		position: relative;
	}
	
	.theme-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.25rem;
		padding: 0.5rem;
		min-width: 2.5rem;
		height: 2.5rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
		background: rgba(255, 255, 255, 0.05);
		border-radius: 0.5rem;
		color: var(--color-text-muted);
		font-size: 0.875rem;
		cursor: pointer;
		transition: all 0.2s;
	}
	
	.theme-btn:hover {
		background: rgba(255, 255, 255, 0.1);
		border-color: rgba(255, 255, 255, 0.2);
		color: var(--color-text-light);
	}
	
	:global(:root.light) .theme-btn {
		border-color: rgba(0, 0, 0, 0.25);
		border-width: 1.5px;
		background: rgba(0, 0, 0, 0.06);
		color: var(--color-text);
	}
	
	:global(:root.light) .theme-btn:hover {
		background: rgba(0, 0, 0, 0.12);
		border-color: rgba(0, 0, 0, 0.4);
		border-width: 1.5px;
		color: var(--color-text-light);
	}
	
	.theme-icon {
		font-size: 1.125rem;
	}
	
	.chevron {
		transition: transform 0.2s;
	}
	
	.chevron.rotated {
		transform: rotate(180deg);
	}
	
	.dropdown {
		position: absolute;
		top: calc(100% + 0.5rem);
		right: 0;
		min-width: 140px;
		background: var(--color-surface);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 0.75rem;
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.4);
		overflow: hidden;
		animation: dropdownIn 0.2s ease-out;
		padding: 0.25rem;
		z-index: 100;
	}
	
	:global(:root.light) .dropdown {
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
	
	.dropdown-item {
		display: flex;
		align-items: center;
		gap: 0.625rem;
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
	
	:global(:root.light) .dropdown-item {
		color: var(--color-text);
	}
	
	.dropdown-item:hover {
		background: rgba(255, 255, 255, 0.08);
	}
	
	:global(:root.light) .dropdown-item:hover {
		background: rgba(0, 0, 0, 0.12);
	}
	
	.dropdown-item.selected {
		background: rgba(99, 102, 241, 0.15);
		color: var(--color-primary-light);
	}
	
	:global(:root.light) .dropdown-item.selected {
		background: rgba(99, 102, 241, 0.2);
		color: var(--color-primary-dark);
	}
	
	.option-icon {
		font-size: 1rem;
	}
	
	.option-label {
		flex: 1;
	}
	
	.check-icon {
		color: var(--color-primary);
	}
</style>
