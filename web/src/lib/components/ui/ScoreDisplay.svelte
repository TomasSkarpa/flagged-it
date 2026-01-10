<script lang="ts">
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	
	export let score = 0;
	export let total = 10;
	export let showPercentage = true;
	export let showProgress = true;
	
	$: currentLocale = $locale;
	$: scoreLabel = t('game.score.label', undefined, currentLocale);
	$: percentage = total > 0 ? Math.round((score / total) * 100) : 0;
	$: percentageClass = percentage >= 80 ? 'text-sage' : percentage >= 60 ? 'text-sunset' : 'text-pin-red';
</script>

<div class="card-stats px-6 py-3">
	<p class="text-xs text-sandy uppercase tracking-wide font-semibold">{scoreLabel}</p>
	<div class="flex items-baseline gap-2">
		<p class="text-2xl font-bold stat-number text-sandy-light leading-none">{score}</p>
		<span class="text-base text-slate-light">/ {total}</span>
		{#if showPercentage}
			<span class="text-base font-bold {percentageClass} stat-number">({percentage}%)</span>
		{/if}
	</div>
	{#if showProgress}
		<div class="mt-2 w-32">
			<div class="w-full h-1.5 bg-white/10 rounded-full overflow-hidden">
				<div 
					class="h-full bg-terracotta transition-all duration-300 rounded-full"
					style="width: {percentage}%"
				></div>
			</div>
		</div>
	{/if}
</div>


