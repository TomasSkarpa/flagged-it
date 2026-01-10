<script>
	export let title = '';
	export let description = '';
	export let icon = '🌍';
	export let selected = false;
	export let difficulty = 'medium';
	
	$: difficultyColors = {
		easy: 'border-sage text-sage',
		medium: 'border-sunset text-sunset',
		hard: 'border-pin-red text-pin-red'
	};
</script>

<button
	class="card-country p-6 text-left transition-all duration-200 {selected ? 'selected' : ''}"
>
	<div class="flex items-start gap-4">
		<div class="text-5xl flex-shrink-0 leading-none">{icon}</div>
		<div class="flex-1">
			<h3 class="text-card-title text-sandy-light font-heading mb-2 leading-tight mt-0">{title}</h3>
			<p class="text-sm text-sandy mb-3">{description}</p>
			<div class="flex items-center gap-2">
				<span class="text-xs px-2 py-1 rounded border {difficultyColors[difficulty] || difficultyColors.medium}">
					{difficulty}
				</span>
			</div>
		</div>
		{#if selected}
			<div class="flex-shrink-0">
				<span class="text-2xl">✓</span>
			</div>
		{/if}
	</div>
</button>

