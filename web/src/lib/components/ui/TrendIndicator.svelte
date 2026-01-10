<script>
	export let value = '';
	export let direction = 'none'; // 'up', 'down', or 'none'
	export let showArrow = true;
	
	$: arrowColor = 'text-sky'; // Blue arrows as shown in screenshot
	$: arrowIcon = direction === 'up' ? '↑' : direction === 'down' ? '↓' : '';
	$: bgClass = 'bg-pin-red/20'; // Light red background for numeric values (as shown in screenshot)
</script>

<div class="flex items-center gap-1.5 px-2 py-1 rounded {bgClass}">
	<span class="text-sm stat-number text-sandy-light">{value}</span>
	{#if showArrow && arrowIcon}
		<span class="{arrowColor} text-sm font-bold" aria-label={direction === 'up' ? 'Higher' : 'Lower'}>{arrowIcon}</span>
	{/if}
</div>

