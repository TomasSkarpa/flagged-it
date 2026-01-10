<script>
	export let correctAnswer = '';
	export let show = false;
</script>

{#if show}
	<div class="fixed inset-0 flex items-center justify-center z-50 pointer-events-none animate-fade-in">
		<div class="bg-pin-red/90 backdrop-blur-sm rounded-card p-8 text-center animate-shake">
			<div class="text-6xl mb-4">❌</div>
			<p class="text-3xl font-bold text-white mb-2">Wrong!</p>
			<p class="text-xl text-white/90">Correct answer: <span class="font-bold">{correctAnswer}</span></p>
		</div>
	</div>
{/if}

