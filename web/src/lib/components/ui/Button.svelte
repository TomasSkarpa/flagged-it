<script>
	// @ts-nocheck
	export let variant = 'primary';
	export let size = 'md';
	export let disabled = false;
	export let type = 'button';
	export let icon = null;
	
	$: buttonClass = `btn-${variant} ${size === 'sm' ? 'px-4 py-2 text-sm' : size === 'lg' ? 'px-8 py-4 text-lg' : ''}`;
</script>

<button
	type="{type}"
	{disabled}
	class={buttonClass}
	aria-disabled={disabled}
	on:click
>
	{#if icon && variant !== 'icon'}
		<span class="inline-block mr-2" aria-hidden="true">{icon}</span>
	{/if}
	<slot />
</button>

<style>
	button {
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}
</style>

