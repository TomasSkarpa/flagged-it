<script lang="ts">
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';

	export let question: string;
	export let showFeedback: boolean = false;
	export let isCorrect: boolean = false;
	export let correctAnswer: string = '';
	export let error: string | null = null;

	$: currentLocale = $locale;
	$: correctText = t('game.correct_short', undefined, currentLocale);
	$: wrongText = t('game.wrong_short', undefined, currentLocale);
</script>

<div class="card-game relative overflow-hidden">
	<h2 class="text-2xl md:text-3xl font-bold text-sandy-light text-center mb-8">
		{question}
	</h2>
	
	<!-- Main content slot (flag, shape, etc.) -->
	<div class="mb-10 flex justify-center">
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
			class="absolute inset-0 flex items-center justify-center rounded-card animate-fade-in z-20
				{isCorrect ? 'bg-success/50' : 'bg-error/50'}"
		>
			<div class="text-center">
				<div class="text-6xl mb-4">{isCorrect ? '✓' : '✗'}</div>
				<p class="text-3xl font-bold text-white">{isCorrect ? correctText : wrongText}</p>
				{#if !isCorrect && correctAnswer}
					<p class="text-xl text-white/90 mt-2">{correctAnswer}</p>
				{/if}
			</div>
		</div>
	{/if}
</div>
