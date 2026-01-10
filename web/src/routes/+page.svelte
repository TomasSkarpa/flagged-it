<script lang="ts">
	import { goto } from '$app/navigation';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	
	// Reactive translations - will update when locale changes
	// Reference $locale to ensure reactivity
	$: currentLocale = $locale;
	$: games = [
		{
			id: 'flag',
			title: t('game.flag.title', undefined, currentLocale),
			description: t('home.game.flag.description', undefined, currentLocale),
			icon: '🚩',
			route: '/flag-game',
			available: true
		},
		{
			id: 'shape',
			title: t('game.shape.title', undefined, currentLocale),
			description: t('home.game.shape.description', undefined, currentLocale),
			icon: '🗺️',
			route: '/shape-game',
			available: true
		},
		{
			id: 'capital',
			title: t('home.game.capital.title', undefined, currentLocale),
			description: t('home.game.capital.description', undefined, currentLocale),
			icon: '🏛️',
			route: '/capital-game',
			available: true
		},
		{
			id: 'higher-lower',
			title: t('game.higher_lower.title', undefined, currentLocale),
			description: t('home.game.higher_lower.description', undefined, currentLocale),
			icon: '↕️',
			route: '/higher-lower',
			available: true
		},
		{
			id: 'hangman',
			title: t('game.hangman.title', undefined, currentLocale),
			description: t('promo.hangman.desc', undefined, currentLocale) || t('game.hangman.title', undefined, currentLocale),
			icon: '🎯',
			route: '/hangman-game',
			available: true
		},
		{
			id: 'worldle',
			title: t('game.guessing.title', undefined, currentLocale),
			description: t('game.worldle.description', undefined, currentLocale) || t('game.guessing.make_guess', undefined, currentLocale),
			icon: '🌍',
			route: '/worldle-game',
			available: true
		},
		{
			id: 'facts',
			title: t('game.facts.title', undefined, currentLocale),
			description: t('home.game.facts.description', undefined, currentLocale) || t('game.facts.description', undefined, currentLocale),
			icon: '📚',
			route: '/facts-game',
			available: true
		}
	];
	
	$: heroTitle = t('home.hero.title', undefined, currentLocale);
	$: heroSubtitle = t('home.hero.subtitle', undefined, currentLocale);
	$: heroDescription = t('home.hero.description', undefined, currentLocale);
	$: featuresTitle = t('home.features.title', undefined, currentLocale);
	$: multipleModesTitle = t('home.features.multiple_modes.title', undefined, currentLocale);
	$: multipleModesDescription = t('home.features.multiple_modes.description', undefined, currentLocale);
	$: regionalFocusTitle = t('home.features.regional.title', undefined, currentLocale);
	$: regionalFocusDescription = t('home.features.regional.description', undefined, currentLocale);
	$: trackProgressTitle = t('home.features.progress.title', undefined, currentLocale);
	$: trackProgressDescription = t('home.features.progress.description', undefined, currentLocale);
	$: playNow = t('home.game.play_now', undefined, currentLocale);
	$: comingSoon = t('home.game.coming_soon', undefined, currentLocale);
	
	function navigateToGame(route: string, available: boolean) {
		if (available) {
			goto(route);
		}
	}
</script>

<svelte:head>
	<title>{heroTitle} - Country Guessing Games</title>
	<meta name="description" content={heroDescription} />
</svelte:head>

<div class="home-page min-h-screen p-4 md:p-8 relative overflow-hidden">
	<!-- World Map Background -->
	<div class="world-map-background">
		<img 
			src="/assets/world_map_silhouette.svg" 
			alt="" 
			class="world-map-svg"
			aria-hidden="true"
		/>
	</div>
	
	<div class="max-w-6xl mx-auto relative z-10">
		<!-- Hero Section -->
		<div class="text-center mb-16">
			<h1 class="text-5xl md:text-7xl font-bold mb-4">
				<span class="mr-2 emoji-blue">🌍</span><span class="gradient-text">{heroTitle}</span>
			</h1>
			<p class="text-xl md:text-2xl text-text-muted mb-2">
				{heroSubtitle}
			</p>
			<p class="text-lg text-text-dark">
				{heroDescription}
			</p>
		</div>
		
		<!-- Games Grid -->
		<div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-12">
			{#each games as game}
				<button
					on:click={() => navigateToGame(game.route, game.available)}
					disabled={!game.available}
					class="card-game text-left transition-all duration-300 hover:shadow-glow hover:-translate-y-2 
						{game.available ? 'cursor-pointer' : 'opacity-50 cursor-not-allowed'}"
				>
					<div class="flex items-start gap-4">
						<div class="text-5xl leading-none">{game.icon}</div>
						<div class="flex-1">
							<h2 class="text-2xl font-bold text-sandy-light mb-2 leading-none mt-0">
								{game.title}
							</h2>
							<p class="text-text-muted mb-4">
								{game.description}
							</p>
							{#if game.available}
								<span class="inline-block px-4 py-2 bg-primary/20 border border-primary rounded-lg text-primary font-semibold">
									{playNow}
								</span>
							{:else}
								<span class="inline-block px-4 py-2 bg-white/5 border border-white/10 rounded-lg text-text-dark font-semibold">
									{comingSoon}
								</span>
							{/if}
						</div>
					</div>
				</button>
			{/each}
		</div>
		
		<!-- Features Section -->
		<div class="card-game text-center">
			<h2 class="text-3xl font-bold text-sandy-light mb-6">{featuresTitle}</h2>
			<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
				<div class="p-4">
					<div class="text-4xl mb-3">🎯</div>
					<h3 class="text-xl font-semibold text-sandy-light mb-2">{multipleModesTitle}</h3>
					<p class="text-text-muted">
						{multipleModesDescription}
					</p>
				</div>
				<div class="p-4">
					<div class="text-4xl mb-3">🌏</div>
					<h3 class="text-xl font-semibold text-sandy-light mb-2">{regionalFocusTitle}</h3>
					<p class="text-text-muted">
						{regionalFocusDescription}
					</p>
				</div>
				<div class="p-4">
					<div class="text-4xl mb-3">📊</div>
					<h3 class="text-xl font-semibold text-sandy-light mb-2">{trackProgressTitle}</h3>
					<p class="text-text-muted">
						{trackProgressDescription}
					</p>
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	.home-page {
		position: relative;
	}
	
	.world-map-background {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		pointer-events: none;
		z-index: 0;
		overflow: hidden;
	}
	
	.world-map-svg {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		width: 120%;
		max-width: 1400px;
		height: auto;
		opacity: 0.15;
		mix-blend-mode: overlay;
		filter: blur(1px);
		transition: opacity 0.3s ease;
	}
	
	/* Adjust blend mode on different backgrounds */
	@media (prefers-color-scheme: dark) {
		.world-map-svg {
			mix-blend-mode: soft-light;
			opacity: 0.12;
		}
	}
	
	/* Ensure content is above background */
	.home-page > :global(.max-w-6xl) {
		position: relative;
		z-index: 1;
	}
	
	/* Subtle animation on hover for interactivity */
	@media (hover: hover) {
		.home-page:hover .world-map-svg {
			opacity: 0.18;
		}
	}
</style>
