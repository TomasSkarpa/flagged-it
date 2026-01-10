<script>
	export let distance = 0; // in km
	
	$: colorClass = distance < 100 ? 'text-sage' : distance < 500 ? 'text-sunset' : 'text-pin-red';
	$: bgClass = distance < 100 ? 'bg-sage/20 border-sage' : distance < 500 ? 'bg-sunset/20 border-sunset' : 'bg-pin-red/20 border-pin-red';
</script>

<div class="inline-flex items-center gap-2 px-4 py-2 rounded-lg border-2 {bgClass} pointer-events-none">
	<span class="text-xl">📍</span>
	<div>
		<span class="text-lg font-bold stat-number {colorClass}">{distance}</span>
		<span class="text-sm text-slate-light ml-1">km away</span>
	</div>
</div>

