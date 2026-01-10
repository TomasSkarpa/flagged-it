<script>
	export let icon = '';
	export let label = '';
	export let value = '';
	export let accent = false;
</script>

<div class="card-stats items-baseline">
	{#if icon}
		<span class="text-2xl opacity-70" role="img" aria-label={label}>{icon}</span>
	{/if}
	<div class="flex-1 flex flex-col">
		<p class="text-xs text-sandy uppercase tracking-wide font-semibold mb-1">{label}</p>
		<p class="text-2xl font-bold stat-number leading-none {accent ? 'text-sunset' : 'text-sandy-light'}">{value}</p>
	</div>
</div>

