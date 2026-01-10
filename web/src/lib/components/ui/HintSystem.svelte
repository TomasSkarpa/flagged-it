<script>
	export let hintLevel = 0; // 0 = no hint, 1 = continent, 2 = region, 3 = proximity
	export let continent = '';
	export let region = '';
	export let proximity = '';
	export let cost = 10;
	
	let revealed = [];
	
	function revealHint(level) {
		if (!revealed.includes(level)) {
			revealed = [...revealed, level];
		}
	}
</script>

<div class="w-full">
	<div class="flex items-center gap-3 mb-4">
		<span class="text-2xl">💡</span>
		<p class="text-xs text-sandy uppercase tracking-wide font-semibold">HINTS</p>
	</div>
	<div class="space-y-2">
		{#if hintLevel >= 1 || revealed.includes(1)}
			<button
				on:click={() => revealHint(1)}
				class="w-full text-left px-4 py-3 rounded-lg border transition-all cursor-pointer flex items-center gap-3 text-base font-medium hint-button"
				class:revealed={revealed.includes(1)}
				disabled={revealed.includes(1)}
			>
				<span class="text-2xl flex-shrink-0">{revealed.includes(1) ? '🌍' : '🌐'}</span>
				<span class="flex-1 text-white">
					{revealed.includes(1) ? `Continent: ${continent}` : 'Reveal Continent'}
				</span>
				{#if !revealed.includes(1)}
					<span class="text-xs text-slate-300 font-normal">({cost} pts)</span>
				{/if}
			</button>
		{/if}
		{#if hintLevel >= 2 || revealed.includes(2)}
			<button
				on:click={() => revealHint(2)}
				class="w-full text-left px-4 py-3 rounded-lg border transition-all cursor-pointer flex items-center gap-3 text-base font-medium hint-button"
				class:revealed={revealed.includes(2)}
				disabled={revealed.includes(2)}
			>
				<span class="text-2xl flex-shrink-0">{revealed.includes(2) ? '📍' : '🗺️'}</span>
				<span class="flex-1 text-white">
					{revealed.includes(2) ? `Region: ${region}` : 'Reveal Region'}
				</span>
				{#if !revealed.includes(2)}
					<span class="text-xs text-slate-300 font-normal">({cost} pts)</span>
				{/if}
			</button>
		{/if}
		{#if hintLevel >= 3 || revealed.includes(3)}
			<button
				on:click={() => revealHint(3)}
				class="w-full text-left px-4 py-3 rounded-lg border transition-all cursor-pointer flex items-center gap-3 text-base font-medium hint-button"
				class:revealed={revealed.includes(3)}
				disabled={revealed.includes(3)}
			>
				<span class="text-2xl flex-shrink-0">{revealed.includes(3) ? '📏' : '📐'}</span>
				<span class="flex-1 text-white">
					{revealed.includes(3) ? `Proximity: ${proximity}` : 'Reveal Proximity'}
				</span>
				{#if !revealed.includes(3)}
					<span class="text-xs text-slate-300 font-normal">({cost} pts)</span>
				{/if}
			</button>
		{/if}
	</div>
</div>

