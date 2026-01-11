<script lang="ts">
	import { locale } from '$lib/stores/locale';
	
	let isOpen = false;
	let dropdownRef: HTMLDivElement;
	
	const languages = locale.getSupportedLocales();
	
	// Language to flag mapping
	const languageFlags: Record<string, string> = {
		'en': '🇬🇧',
		'es': '🇪🇸',
		'fr': '🇫🇷',
		'de': '🇩🇪',
		'nl': '🇳🇱',
		'nb': '🇳🇴',
		'da': '🇩🇰',
		'sv': '🇸🇪',
		'fi': '🇫🇮',
		'pt': '🇵🇹',
		'tr': '🇹🇷',
		'ro': '🇷🇴',
		'hu': '🇭🇺',
		'hr': '🇭🇷',
		'cs': '🇨🇿',
		'sk': '🇸🇰',
		'pl': '🇵🇱',
		'it': '🇮🇹',
		'id': '🇮🇩',
		'ms': '🇲🇾',
		'fil': '🇵🇭',
		'sw': '🇰🇪',
		'vi': '🇻🇳',
		'ru': '🇷🇺',
		'zh': '🇨🇳',
		'ko': '🇰🇷',
		'ja': '🇯🇵',
		'ar': '🇸🇦',
		'hi': '🇮🇳',
		'th': '🇹🇭',
		'uk': '🇺🇦',
		'he': '🇮🇱',
		'el': '🇬🇷'
	};
	
	function toggleDropdown() {
		isOpen = !isOpen;
	}
	
	function selectLanguage(code: string) {
		locale.set(code);
		isOpen = false;
	}
	
	function handleClickOutside(event: MouseEvent) {
		if (dropdownRef && !dropdownRef.contains(event.target as Node)) {
			isOpen = false;
		}
	}
	
	function getCurrentLanguage() {
		return languages.find(l => l.code === $locale) || languages[0];
	}
	
	$: currentLang = getCurrentLanguage();
</script>

<svelte:window on:click={handleClickOutside} />

<div class="language-selector" bind:this={dropdownRef}>
	<button 
		class="language-btn"
		on:click|stopPropagation={toggleDropdown}
		aria-label="Select language"
		aria-expanded={isOpen}
	>
		<span class="flag">{languageFlags[$locale] || '🌐'}</span>
		<span class="lang-code">{$locale.toUpperCase()}</span>
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
			<div class="dropdown-header">Select Language</div>
			<div class="dropdown-list">
				{#each languages as lang}
					<button
						class="dropdown-item"
						class:selected={$locale === lang.code}
						on:click|stopPropagation={() => selectLanguage(lang.code)}
					>
						<span class="flag">{languageFlags[lang.code] || '🌐'}</span>
						<span class="lang-name">{lang.name}</span>
						{#if $locale === lang.code}
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
	.language-selector {
		position: relative;
	}
	
	.language-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.375rem;
		padding: 0.5rem;
		min-width: 2.5rem;
		height: 2.5rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
		background: rgba(255, 255, 255, 0.05);
		border-radius: 0.5rem;
		color: var(--color-text-muted);
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
	}
	
	.language-btn:hover {
		background: rgba(255, 255, 255, 0.1);
		border-color: rgba(255, 255, 255, 0.2);
		color: var(--color-text-light);
	}
	
	:global(:root.light) .language-btn {
		border-color: rgba(0, 0, 0, 0.25);
		border-width: 1.5px;
		background: rgba(0, 0, 0, 0.06);
		color: var(--color-text);
	}
	
	:global(:root.light) .language-btn:hover {
		background: rgba(0, 0, 0, 0.12);
		border-color: rgba(0, 0, 0, 0.4);
		border-width: 1.5px;
		color: var(--color-text-light);
	}
	
	.flag {
		font-size: 1rem;
	}
	
	.lang-code {
		display: none;
	}
	
	@media (min-width: 768px) {
		.language-btn {
			padding: 0.5rem 0.75rem;
			min-width: auto;
		}
		.lang-code {
			display: inline;
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
		min-width: 200px;
		max-height: 320px;
		background: var(--color-surface);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 0.75rem;
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.4);
		overflow: hidden;
		animation: dropdownIn 0.2s ease-out;
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
	
	.dropdown-header {
		padding: 0.75rem 1rem;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--color-text-muted);
		border-bottom: 1px solid rgba(255, 255, 255, 0.08);
	}
	
	:global(:root.light) .dropdown-header {
		border-bottom-color: rgba(0, 0, 0, 0.2);
		border-bottom-width: 1.5px;
		color: var(--color-text);
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
	
	:global(:root.light) .dropdown-list::-webkit-scrollbar-thumb {
		background: rgba(0, 0, 0, 0.2);
	}
	
	.dropdown-item {
		display: flex;
		align-items: center;
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
	
	.lang-name {
		flex: 1;
	}
	
	.check-icon {
		color: var(--color-primary);
	}
</style>
