<script>
	export let points = 0;
	export let show = false;
</script>

{#if show}
	<div class="fixed inset-0 flex items-center justify-center z-50 pointer-events-none animate-fade-in">
		<div class="bg-sage/90 backdrop-blur-sm rounded-card p-8 text-center animate-bounce-gentle">
			<div class="text-6xl mb-4">✅</div>
			<p class="text-3xl font-bold text-white mb-2">Correct!</p>
			<p class="text-xl text-white/90 stat-number">+{points} points</p>
		</div>
	</div>
{/if}

