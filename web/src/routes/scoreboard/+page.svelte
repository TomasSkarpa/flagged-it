<script lang="ts">
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';

	$: currentLocale = $locale;
	$: title = t('scoreboard.title', undefined, currentLocale);
	$: description = t('scoreboard.description', undefined, currentLocale);
	$: comingSoon = t('scoreboard.coming_soon', undefined, currentLocale);
</script>

<svelte:head>
	<title>{title} - Flagged It</title>
</svelte:head>

<div class="min-h-screen p-4 md:p-8">
	<div class="max-w-4xl mx-auto">
		<div class="card-game text-center py-16">
			<div class="text-6xl mb-6">🏆</div>
			<h1 class="text-4xl font-bold text-sandy-light mb-4">{title}</h1>
			<p class="text-xl text-text-muted mb-8">
				{description}
			</p>
			<div class="inline-block px-6 py-3 bg-white/5 border border-white/10 rounded-lg text-text-dark font-semibold">
				{comingSoon}
			</div>
		</div>
	</div>
</div>
