<script>
	import { resolveAssetUrl } from '$lib/api/config';

	export let flags = [];
	export let selected = [];
	export let searchQuery = '';
	
	$: filteredFlags = flags.filter(flag => 
		flag.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
		flag.code.toLowerCase().includes(searchQuery.toLowerCase())
	);
	
	function toggleSelection(code) {
		if (selected.includes(code)) {
			selected = selected.filter(c => c !== code);
		} else {
			selected = [...selected, code];
		}
	}
</script>

<div class="space-y-4">
	<!-- Search bar - Darker for depth -->
	<input
		type="text"
		bind:value={searchQuery}
		placeholder="Search countries..."
		class="input-base w-full"
	/>
	
	<!-- Grid -->
	<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
		{#each filteredFlags as flag}
			<button
				on:click={() => toggleSelection(flag.code)}
				class="relative card-country p-3 text-center transition-all duration-200 flag-card {selected.includes(flag.code) ? 'selected' : ''}"
			>
				<!-- Larger flag area - no inner box -->
				<div class="mb-2 flex items-center justify-center h-24">
					<img src={resolveAssetUrl(flag.flagUrl)} alt={`${flag.name} flag`} class="max-w-full max-h-full object-contain rounded-sm" />
				</div>
				<p class="text-sm text-white font-medium">{flag.name}</p>
				{#if selected.includes(flag.code)}
					<div class="absolute top-2 right-2 bg-white rounded p-1 shadow-lg flex items-center justify-center w-5 h-5">
						<span class="text-sage text-xs font-bold leading-none">✓</span>
					</div>
				{/if}
			</button>
		{/each}
	</div>
</div>

