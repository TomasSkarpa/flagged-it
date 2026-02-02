<script lang="ts">
	import { theme, type Theme } from '$lib/stores/theme';
	import { activeDropdown, toggleDropdown as toggleDropdownStore, closeDropdown } from '$lib/stores/dropdown';
	import { onMount, onDestroy } from 'svelte';
	
	let dropdownRef: HTMLDivElement;
	let isHovered = false;
	let cycleTimeout: ReturnType<typeof setTimeout> | null = null;
	let displayedIconIndex = 0;
	let isAnimating = false;
	
	const themes: { value: Theme; label: string; icon: string }[] = [
		{ value: 'light', label: 'Light', icon: '☀️' },
		{ value: 'dark', label: 'Dark', icon: '🌙' },
		{ value: 'system', label: 'System', icon: '💻' }
	];
	
	function toggleDropdown() {
		toggleDropdownStore('theme');
	}
	
	function selectTheme(value: Theme) {
		theme.set(value);
		closeDropdown();
		// Reset to show current theme icon after selection
		const currentIndex = themes.findIndex(t => t.value === value);
		if (currentIndex !== -1) {
			displayedIconIndex = currentIndex;
		}
	}
	
	function handleClickOutside(event: MouseEvent) {
		if (dropdownRef && !dropdownRef.contains(event.target as Node)) {
			closeDropdown();
		}
	}
	
	function getCurrentIcon() {
		const current = themes.find(t => t.value === $theme);
		return current?.icon || '🌙';
	}
	
	function getDisplayedIcon() {
		return themes[displayedIconIndex]?.icon || getCurrentIcon();
	}
	
	function getRandomDelay(): number {
		// Random delay between 7-30 seconds (7000-30000ms)
		return Math.floor(Math.random() * 23000) + 7000;
	}
	
	function showAllThemes() {
		if (isOpen || isHovered) return;
		
		const currentThemeIndex = themes.findIndex(t => t.value === $theme);
		let currentCycleIndex = 0;
		
		// Function to show next theme in the cycle
		const showNext = () => {
			if (isOpen || isHovered) {
				// Reset to current theme if interrupted
				if (currentThemeIndex !== -1) {
					displayedIconIndex = currentThemeIndex;
					isAnimating = false;
				}
				return;
			}
			
			// Set the icon index first
			displayedIconIndex = currentCycleIndex;
			
			// Safari: Use requestAnimationFrame to ensure proper class application
			// For the first icon, ensure animation starts properly
			requestAnimationFrame(() => {
				requestAnimationFrame(() => {
					isAnimating = true;
				});
			});
			
			currentCycleIndex++;
			
			if (currentCycleIndex < themes.length) {
				// Continue cycling through all themes
				setTimeout(() => {
					isAnimating = false;
					setTimeout(showNext, 100); // Delay between transitions
				}, 800);
			} else {
				// After showing all themes (0, 1, 2), return to current theme
				// At this point, displayedIconIndex is 2 (the last icon shown)
				setTimeout(() => {
					if (currentThemeIndex !== -1) {
						// Ensure we're showing the current theme icon
						// If it's already correct, we can just clear animation
						// Otherwise, set it first to ensure it's visible
						if (displayedIconIndex !== currentThemeIndex) {
							displayedIconIndex = currentThemeIndex;
						}
						// Clear animation after ensuring icon is set
						// Use a small delay to ensure the index is applied
						setTimeout(() => {
							isAnimating = false;
							scheduleNextShowcase();
						}, 10);
					} else {
						isAnimating = false;
						scheduleNextShowcase();
					}
				}, 800);
			}
		};
		
		// Start the cycle - ensure first icon (sun/light at index 0) is shown properly
		// Clear any existing animation state first
		isAnimating = false;
		// Small delay to ensure Safari processes the state change, then start cycle
		// The first call to showNext() will show index 0 (sun), then cycle through 1 (moon), 2 (computer)
		setTimeout(() => {
			showNext();
		}, 100);
	}
	
	function scheduleNextShowcase() {
		if (cycleTimeout) clearTimeout(cycleTimeout);
		if (isOpen || isHovered) return;
		
		const delay = getRandomDelay();
		cycleTimeout = setTimeout(() => {
			if (!isOpen && !isHovered) {
				showAllThemes();
			}
		}, delay);
	}
	
	function startCycling() {
		// Don't start if already cycling or if conditions aren't met
		if (cycleTimeout || isOpen || isHovered) return;
		
		// Find current theme index to start from
		const currentIndex = themes.findIndex(t => t.value === $theme);
		if (currentIndex !== -1) {
			displayedIconIndex = currentIndex;
		}
		
		// Start cycling after initial delay (4 seconds)
		setTimeout(() => {
			if (isOpen || isHovered) return; // Double check conditions
			isCyclingActive = true;
			scheduleNextShowcase();
		}, 4000);
	}
	
	// Track if we're actively cycling to prevent reactive updates
	let isCyclingActive = false;
	
	function stopCycling() {
		isCyclingActive = false;
		if (cycleTimeout) {
			clearTimeout(cycleTimeout);
			cycleTimeout = null;
		}
		// Reset to current theme icon
		const currentIndex = themes.findIndex(t => t.value === $theme);
		if (currentIndex !== -1) {
			displayedIconIndex = currentIndex;
		}
	}
	
	function handleMouseEnter() {
		isHovered = true;
		stopCycling();
	}
	
	function handleMouseLeave() {
		isHovered = false;
		// Restart cycling after a delay when mouse leaves
		setTimeout(() => {
			if (!isOpen && !isHovered) {
				startCycling();
			}
		}, 1000);
	}
	
	$: isOpen = $activeDropdown === 'theme';
	
	// Update displayed icon when theme changes (only if not cycling)
	$: {
		if (!isCyclingActive && !cycleTimeout && !isAnimating && !isHovered) {
			const currentIndex = themes.findIndex(t => t.value === $theme);
			if (currentIndex !== -1 && displayedIconIndex !== currentIndex) {
				displayedIconIndex = currentIndex;
			}
		}
	}
	
	// Pause cycling when dropdown opens
	let restartTimeout: ReturnType<typeof setTimeout> | null = null;
	$: if (isOpen) {
		stopCycling();
		if (restartTimeout) clearTimeout(restartTimeout);
	} else {
		// Restart cycling when dropdown closes (if not hovered)
		if (restartTimeout) clearTimeout(restartTimeout);
		restartTimeout = setTimeout(() => {
			if (!isOpen && !isHovered) {
				startCycling();
			}
		}, 500);
	}
	
	onMount(() => {
		const currentIndex = themes.findIndex(t => t.value === $theme);
		if (currentIndex !== -1) {
			displayedIconIndex = currentIndex;
		}
		startCycling();
	});
	
	onDestroy(() => {
		stopCycling();
		if (restartTimeout) clearTimeout(restartTimeout);
	});
</script>

<svelte:window on:click={handleClickOutside} />

<div class="theme-toggle" bind:this={dropdownRef}>
	<button 
		class="theme-btn"
		on:click|stopPropagation={toggleDropdown}
		on:mouseenter={handleMouseEnter}
		on:mouseleave={handleMouseLeave}
		aria-label="Select theme"
		aria-expanded={isOpen}
	>
		<span class="theme-icon-wrapper">
			{#each themes as themeOption, index}
				<span 
					class="theme-icon" 
					class:active={displayedIconIndex === index}
					class:animating={isAnimating && displayedIconIndex === index}
				>
					{themeOption.icon}
				</span>
			{/each}
		</span>
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
	
	.theme-icon-wrapper {
		position: relative;
		display: inline-block;
		width: 1.125rem;
		height: 1.125rem;
		overflow: hidden;
		/* Safari: Force hardware acceleration */
		transform: translateZ(0);
		-webkit-transform: translateZ(0);
	}
	
	.theme-icon {
		font-size: 1.125rem;
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0;
		transform: translateX(100%) translateZ(0);
		-webkit-transform: translateX(100%) translateZ(0);
		transition: opacity 0.7s ease, transform 0.7s ease;
		pointer-events: none;
		/* Safari: Force hardware acceleration and prevent flickering */
		will-change: transform, opacity;
		backface-visibility: hidden;
		-webkit-backface-visibility: hidden;
	}
	
	.theme-icon.active {
		opacity: 1;
		transform: translateX(0) translateZ(0);
		-webkit-transform: translateX(0) translateZ(0);
		pointer-events: auto;
	}
	
	.theme-icon.animating {
		animation: iconSlide 0.7s ease;
		/* Safari: Ensure animation uses hardware acceleration */
		will-change: transform, opacity;
	}
	
	@keyframes iconSlide {
		0% {
			transform: translateX(-100%) translateZ(0);
			-webkit-transform: translateX(-100%) translateZ(0);
			opacity: 0;
		}
		100% {
			transform: translateX(0) translateZ(0);
			-webkit-transform: translateX(0) translateZ(0);
			opacity: 1;
		}
	}
	
	/* Safari-specific fixes */
	@supports (-webkit-appearance: none) {
		.theme-icon {
			/* Safari: Ensure proper stacking */
			z-index: 1;
		}
		
		.theme-icon.active {
			z-index: 2;
		}
		
		.theme-icon.animating {
			z-index: 3;
		}
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
