<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { onMount } from 'svelte';
	import ScoreDisplay from '$lib/components/ui/ScoreDisplay.svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import { triggerConfetti } from '$lib/utils/confetti';

	export let score: number;
	export let totalRounds: number;
	export let title: string = 'Game Complete!';
	export let emoji: string = '🎉';
	export let playAgainText: string = '';
	export let showHomeButton: boolean = false;
	export let excellentThreshold: number = 0.8;
	export let goodThreshold: number = 0.6;
	export let excellentMessage: string = '';
	export let goodMessage: string = '';
	export let encourageMessage: string = '';

	$: currentLocale = $locale;
	$: finalPlayAgainText = playAgainText || t('game.over.play_again', undefined, currentLocale);
	$: finalExcellentMessage = excellentMessage || t('game.over.excellent', undefined, currentLocale);
	$: finalGoodMessage = goodMessage || t('game.over.good', undefined, currentLocale);
	$: finalEncourageMessage = encourageMessage || t('game.over.encourage', undefined, currentLocale);
	$: homeText = t('common.home', undefined, currentLocale);

	const dispatch = createEventDispatcher<{
		playAgain: void;
		goHome: void;
	}>();

	$: percentage = score / totalRounds;
	$: message = percentage >= excellentThreshold 
		? finalExcellentMessage 
		: percentage >= goodThreshold 
			? finalGoodMessage 
			: finalEncourageMessage;

	// Trigger confetti for perfect scores (10/10) or excellent scores (80%+)
	onMount(() => {
		if (percentage >= excellentThreshold) {
			// Small delay to ensure component is mounted
			setTimeout(() => {
				triggerConfetti({
					particleCount: percentage === 1 ? 100 : 70,
					spread: 70,
					origin: { x: 0.5, y: 0.3 },
					duration: percentage === 1 ? 4000 : 3000
				});
			}, 100);
		}
	});
</script>

<div class="text-center">
	<div class="card-game max-w-2xl mx-auto">
		<div class="text-6xl mb-6">{emoji}</div>
		<h2 class="text-4xl font-bold text-sandy-light mb-4">{title}</h2>
		<div class="mb-8">
			<ScoreDisplay {score} total={totalRounds} showPercentage={true} showProgress={true} />
		</div>
		<p class="text-xl text-text-muted mb-8">
			{message}
		</p>
		<div class="flex flex-col sm:flex-row gap-4 justify-center">
			<button 
				class="btn-primary px-12 py-4 text-xl font-bold"
				on:click={() => dispatch('playAgain')}
			>
				{finalPlayAgainText}
			</button>
			{#if showHomeButton}
				<button 
					class="btn-secondary px-12 py-4 text-xl font-bold"
					on:click={() => dispatch('goHome')}
				>
					{homeText}
				</button>
			{/if}
		</div>
	</div>
</div>
