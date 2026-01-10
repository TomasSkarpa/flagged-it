<script>
	export let icon = '🧭';
	export let title = 'No data yet';
	export let message = 'Get started by exploring the world!';
	export let actionLabel = 'Start Exploring';
	export let onAction = null;
</script>

<div class="flex flex-col items-center justify-center py-16 px-4 text-center">
	<div class="text-6xl mb-6 animate-pulse-slow">{icon}</div>
	<h2 class="text-section text-sandy font-heading mb-3">{title}</h2>
	<p class="text-lg text-slate-light mb-8 max-w-md">{message}</p>
	{#if onAction}
		<button on:click={onAction} class="btn-primary">
			{actionLabel}
		</button>
	{/if}
</div>

