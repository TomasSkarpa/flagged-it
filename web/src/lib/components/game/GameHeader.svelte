<script lang="ts">
	import ScoreDisplay from '$lib/components/ui/ScoreDisplay.svelte';
	import ProgressBar from '$lib/components/ui/ProgressBar.svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';

	export let score: number;
	export let total: number;
	export let currentRound: number;
	export let totalRounds: number;
	export let showProgressBar: boolean = true;
	
	// Reactive translations - will update when locale changes
	$: currentLocale = $locale;
	$: roundText = t('game.round.progress', [currentRound, totalRounds], currentLocale);
</script>

<div class="mb-6">
	<div class="flex flex-col md:flex-row items-center justify-between gap-4 mb-4">
		<ScoreDisplay {score} {total} showPercentage={false} showProgress={false} />
		<div class="text-lg text-text-light font-semibold stat-number">
			{roundText}
		</div>
	</div>
	{#if showProgressBar}
		<ProgressBar current={total} total={totalRounds} showLabel={false} />
	{/if}
</div>
