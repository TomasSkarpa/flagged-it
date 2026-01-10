<script>
	import DistanceIndicator from './DistanceIndicator.svelte';
	
	export let score = 0;
	export let total = 10;
	export let distance = 0;
	export let showPercentage = true;
	
	$: percentage = total > 0 ? Math.round((score / total) * 100) : 0;
	$: percentageClass = percentage >= 80 ? 'text-sage' : percentage >= 60 ? 'text-sunset' : 'text-pin-red';
</script>

<div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 p-4 rounded-lg backdrop-blur-sm border-2 border-white/20 bg-white/10">
	<div class="flex-1">
		<p class="text-xs text-sandy uppercase tracking-wide font-semibold mb-1">Score</p>
		<div class="flex items-baseline gap-2">
			<span class="text-3xl font-bold stat-number text-sandy-light">{score}</span>
			<span class="text-lg text-slate-light">/ {total}</span>
			{#if showPercentage}
				<span class="text-lg font-bold {percentageClass} stat-number">({percentage}%)</span>
			{/if}
		</div>
	</div>
	<div class="flex-shrink-0">
		<DistanceIndicator distance={distance} />
	</div>
</div>

