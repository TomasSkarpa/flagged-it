<script>
	import { locale } from '$lib/stores/locale';
	
	let selectedLocale = $locale;
	let isOpen = false;
	
	$: supportedLocales = locale.getSupportedLocales();
	$: currentLocaleName = supportedLocales.find(l => l.code === selectedLocale)?.name || 'English';
	
	function handleSubmit() {
		locale.set(selectedLocale);
		isOpen = false;
		// In a real app, you might want to reload the page or trigger a language change event
		// window.location.reload();
	}
	
	function handleCancel() {
		selectedLocale = $locale;
		isOpen = false;
	}
	
	function toggleDropdown() {
		isOpen = !isOpen;
		if (!isOpen) {
			selectedLocale = $locale; // Reset on close
		}
	}
</script>

<div class="relative">
	<button
		type="button"
		on:click={toggleDropdown}
		class="flex items-center justify-between w-full px-4 py-3 rounded-lg border-2 border-white/20 bg-white/10 text-sandy-light hover:border-sky hover:bg-white/15 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-sky focus:ring-offset-2 focus:ring-offset-ocean-dark"
		aria-label="Select language"
		aria-expanded={isOpen}
		aria-haspopup="listbox"
	>
		<div class="flex items-center gap-2">
			<span class="text-xl">🌐</span>
			<span class="font-medium">{currentLocaleName}</span>
		</div>
		<svg 
			class="w-5 h-5 transition-transform duration-200 {isOpen ? 'rotate-180' : ''}" 
			fill="none" 
			stroke="currentColor" 
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
		</svg>
	</button>
	
	{#if isOpen}
		<!-- Backdrop -->
		<div 
			class="fixed inset-0 z-40" 
			on:click={handleCancel}
			on:keydown={(e) => e.key === 'Escape' && handleCancel()}
			role="button"
			tabindex="-1"
			aria-label="Close language selector"
		></div>
		
		<!-- Dropdown -->
		<div 
			class="absolute z-50 w-full mt-2 rounded-lg border-2 border-white/20 bg-ocean-dark backdrop-blur-md shadow-lg overflow-hidden"
			role="listbox"
		>
			<div class="max-h-64 overflow-y-auto">
				{#each supportedLocales as loc}
					<button
						type="button"
						on:click={() => { selectedLocale = loc.code; }}
						class="w-full px-4 py-3 text-left hover:bg-white/10 transition-colors flex items-center justify-between {selectedLocale === loc.code ? 'bg-sky/20 border-l-4 border-sky' : ''}"
						role="option"
						aria-selected={selectedLocale === loc.code}
					>
						<span class="text-sandy-light font-medium">{loc.name}</span>
						{#if selectedLocale === loc.code}
							<span class="text-sky text-lg">✓</span>
						{/if}
					</button>
				{/each}
			</div>
			
			<!-- Action buttons -->
			<div class="border-t border-white/20 p-3 flex gap-2">
				<button 
					type="button"
					on:click={handleCancel}
					class="flex-1 px-4 py-2 rounded-button font-semibold transition-all duration-200 bg-transparent border border-sage text-sage hover:bg-sage/20 focus:outline-none focus:ring-2 focus:ring-sage"
				>
					Cancel
				</button>
				<button 
					type="button"
					on:click={handleSubmit}
					class="flex-1 px-4 py-2 rounded-button font-semibold transition-all duration-200 bg-terracotta text-white hover:bg-terracotta-dark focus:outline-none focus:ring-2 focus:ring-terracotta"
				>
					Apply
				</button>
			</div>
		</div>
	{/if}
</div>

<style>
	/* Smooth scrollbar styling */
	:global(.max-h-64::-webkit-scrollbar) {
		width: 8px;
	}
	
	:global(.max-h-64::-webkit-scrollbar-track) {
		background: rgba(255, 255, 255, 0.05);
		border-radius: 4px;
	}
	
	:global(.max-h-64::-webkit-scrollbar-thumb) {
		background: rgba(255, 255, 255, 0.2);
		border-radius: 4px;
	}
	
	:global(.max-h-64::-webkit-scrollbar-thumb:hover) {
		background: rgba(255, 255, 255, 0.3);
	}
</style>

