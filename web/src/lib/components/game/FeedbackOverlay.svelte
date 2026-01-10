<script lang="ts">
	export let show: boolean = false;
	export let isCorrect: boolean = false;
	export let correctAnswer: string = '';
	export let correctEmoji: string = '✓';
	export let wrongEmoji: string = '✗';
	export let correctText: string = 'Correct!';
	export let wrongText: string = 'Wrong!';
</script>

{#if show}
	<div 
		class="absolute inset-0 flex items-center justify-center rounded-card animate-fade-in z-20
			{isCorrect ? 'bg-success/50' : 'bg-error/50'}"
	>
		<div class="text-center">
			<div class="text-6xl mb-4">{isCorrect ? correctEmoji : wrongEmoji}</div>
			<p class="text-3xl font-bold text-white">{isCorrect ? correctText : wrongText}</p>
			{#if !isCorrect && correctAnswer}
				<p class="text-xl text-white/90 mt-2">{correctAnswer}</p>
			{/if}
		</div>
	</div>
{/if}
