<script>
	export let checked = false;
	export let label = '';
	let id = `toggle-${Math.random().toString(36).substr(2, 9)}`;
</script>

<label for={id} class="flex items-center gap-2 cursor-pointer">
	<div class="relative inline-flex items-center">
		<input
			type="checkbox"
			{id}
			bind:checked
			class="sr-only peer"
		/>
		<div class="w-14 h-7 bg-white/15 rounded-full peer-checked:bg-terracotta transition-all duration-200 peer-focus:ring-2 peer-focus:ring-terracotta peer-focus:ring-offset-2 border-2 border-white/20 peer-checked:border-terracotta"></div>
		<div class="absolute left-1 top-1 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-200 peer-checked:translate-x-7 peer-checked:bg-sandy-light"></div>
	</div>
	{#if label}
		<span class="text-sandy-light text-sm font-medium">{label}</span>
	{/if}
</label>

