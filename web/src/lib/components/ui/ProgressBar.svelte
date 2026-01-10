<script>
	export let current = 0;
	export let total = 10;
	export let showLabel = true;
	export let label = '';
	export let color = 'terracotta';
	
	$: percentage = total > 0 ? Math.min((current / total) * 100, 100) : 0;
	$: colorClass = {
		terracotta: 'bg-terracotta',
		sage: 'bg-sage',
		sunset: 'bg-sunset',
		sky: 'bg-sky'
	}[color] || 'bg-terracotta';
</script>

<div class="w-full">
	{#if showLabel && label}
		<div class="flex justify-between items-center mb-2">
			<span class="text-sm text-sandy-light font-medium">{label}</span>
			<span class="text-sm text-sandy stat-number">{current} / {total}</span>
		</div>
	{/if}
	<div class="w-full h-3 bg-white/10 rounded-full overflow-hidden border border-white/20">
		<div 
			class="h-full {colorClass} transition-all duration-300 ease-out rounded-full"
			style="width: {percentage}%"
		></div>
	</div>
</div>


