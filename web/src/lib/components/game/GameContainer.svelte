<script lang="ts">
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';

	export let question: string;
	export let showFeedback: boolean = false;
	export let isCorrect: boolean = false;
	export let correctAnswer: string = '';
	export let customMessage: string = ''; // Custom message that replaces the default correct/wrong text
	export let error: string | null = null;

	$: currentLocale = $locale;
	$: correctText = t('game.correct_short', undefined, currentLocale);
	$: displayMessage = customMessage || (isCorrect ? correctText : (correctAnswer ? `The correct answer is: ${correctAnswer}` : 'Incorrect'));
</script>

<div class="card-game relative overflow-hidden">
	<!-- Optional header slot for additional info (timer, game type, etc.) -->
	{#if $$slots.header}
		<div class="pt-2 pb-1">
			<slot name="header" />
		</div>
	{/if}
	
	<h2 class="text-xl md:text-2xl font-bold text-sandy-light text-center mb-3 mt-2">
		{question}
	</h2>
	
	<!-- Main content slot (flag, shape, etc.) -->
	<div class="mb-3 flex justify-center">
		<slot name="content" />
	</div>
	
	<!-- Answer buttons slot -->
	<slot name="answers" />
	
	{#if error}
		<div class="mt-6 p-4 bg-error/20 border border-error rounded-lg">
			<p class="text-error font-semibold text-center">{error}</p>
		</div>
	{/if}

	<!-- Feedback Overlay -->
	{#if showFeedback}
		<div 
			class="feedback-overlay absolute inset-0 flex items-center justify-center rounded-card z-20
				{isCorrect ? 'bg-success/50' : 'bg-error/50'}"
		>
			<div class="text-center">
				<div class="text-6xl mb-4">{isCorrect ? '✓' : '✗'}</div>
				<p class="text-3xl font-bold text-white">{displayMessage}</p>
				{#if !isCorrect && correctAnswer && !customMessage}
					<p class="text-xl text-white/90 mt-2">{correctAnswer}</p>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	.card-game {
		min-height: 550px;
	}
	
	@media (min-height: 600px) {
		.card-game {
			min-height: 65vh;
		}
	}
	
	@media (min-height: 900px) {
		.card-game {
			min-height: 60vh;
		}
	}
	
	.card-game h2 {
		margin-top: 0.5rem !important;
	}
	
	.feedback-overlay {
		animation: fadeIn 0.2s ease-in;
	}
	
	@keyframes fadeIn {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}
</style>
