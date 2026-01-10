<script>
	import { onDestroy } from 'svelte';
	
	export let message = '';
	export let type = 'info';
	export let show = false;
	export let duration = 3000;
	
	let timeoutId;
	
	$: if (show) {
		if (timeoutId) clearTimeout(timeoutId);
		timeoutId = setTimeout(() => {
			show = false;
		}, duration);
	}
	
	// Cleanup on destroy
	onDestroy(() => {
		if (timeoutId) clearTimeout(timeoutId);
	});
	
	$: typeClasses = {
		info: 'bg-sky/90 text-white',
		success: 'bg-sage/90 text-white',
		warning: 'bg-sunset/90 text-white',
		error: 'bg-pin-red/90 text-white'
	};
</script>

{#if show}
	<div class="fixed top-4 right-4 z-50 animate-slide-up">
		<div class="px-6 py-4 rounded-lg shadow-lg backdrop-blur-sm border-2 border-white/20 {typeClasses[type] || typeClasses.info}">
			<p class="font-medium">{message}</p>
		</div>
	</div>
{/if}

