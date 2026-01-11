<script lang="ts">
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
	
	// Initialize with fallbacks to ensure content is always present for SEO
	$: heroTitle = t('home.hero.title', undefined, currentLocale) || 'Flagged It';
	$: heroSubtitle = t('home.hero.subtitle', undefined, currentLocale);
	$: heroDescription = t('home.hero.description', undefined, currentLocale);
	$: featuresTitle = t('home.features.title', undefined, currentLocale) || 'Features';
	$: multipleModesTitle = t('home.features.multiple_modes.title', undefined, currentLocale) || 'Multiple Game Modes';
	$: multipleModesDescription = t('home.features.multiple_modes.description', undefined, currentLocale) || 'Enjoy various game modes including flag guessing, shape identification, capitals, and more.';
	$: regionalFocusTitle = t('home.features.regional.title', undefined, currentLocale) || 'Regional Focus';
	$: regionalFocusDescription = t('home.features.regional.description', undefined, currentLocale) || 'Choose specific regions to focus your learning on different parts of the world.';
	$: trackProgressTitle = t('home.features.progress.title', undefined, currentLocale) || 'Track Progress';
	$: trackProgressDescription = t('home.features.progress.description', undefined, currentLocale) || 'Monitor your learning progress and improve your knowledge of world geography.';
	$: playNow = t('home.game.play_now', undefined, currentLocale) || 'Play Now';
	$: comingSoon = t('home.game.coming_soon', undefined, currentLocale) || 'Coming Soon';
	
</script>

<svelte:head>
	<title>{heroTitle} - Country Guessing Games</title>
	<meta name="description" content={heroDescription} />
</svelte:head>

<div class="home-page min-h-screen p-4 md:p-8 relative overflow-hidden">
	<!-- World Map Background -->
	<div class="world-map-background">
		<div class="world-map-overlay"></div>
		<img 
			src="/assets/world_map_silhouette.svg" 
			alt="" 
			class="world-map-svg"
			aria-hidden="true"
		/>
	</div>
	
	<div class="max-w-6xl mx-auto relative z-10">
		<!-- Hero Section -->
		<header class="text-center mb-16">
			<h1 class="text-5xl md:text-7xl font-bold mb-4">
				<span class="mr-2 emoji-blue" aria-hidden="true">🌍</span>
				<span class="gradient-text">{heroTitle || 'Flagged It'}</span>
			</h1>
			{#if heroSubtitle}
				<p class="text-xl md:text-2xl text-text-muted mb-2">
					{heroSubtitle}
				</p>
			{/if}
			{#if heroDescription}
				<p class="text-lg text-text-dark">
					{heroDescription}
				</p>
			{/if}
		</header>
		
		<!-- Games Grid -->
		<div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-12">
			{#each games as game}
				{#if game.available}
					<a
						href={game.route}
						class="card-game text-left transition-all duration-300 hover:shadow-glow hover:-translate-y-2 cursor-pointer block no-underline"
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
								<span class="inline-block px-4 py-2 bg-primary/20 border border-primary rounded-lg text-primary font-semibold">
									{playNow}
								</span>
							</div>
						</div>
					</a>
				{:else}
					<div
						class="card-game text-left opacity-50 cursor-not-allowed"
						aria-disabled="true"
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
								<span class="inline-block px-4 py-2 bg-white/5 border border-white/10 rounded-lg text-text-dark font-semibold">
									{comingSoon}
								</span>
							</div>
						</div>
					</div>
				{/if}
			{/each}
		</div>
		
		<!-- Features Section -->
		<section class="card-game text-center mb-12">
			<h2 class="text-3xl font-bold text-sandy-light mb-6">{featuresTitle || 'Features'}</h2>
			<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
				<article class="p-4">
					<div class="text-4xl mb-3" aria-hidden="true">🎯</div>
					<h3 class="text-xl font-semibold text-sandy-light mb-2">{multipleModesTitle || 'Multiple Game Modes'}</h3>
					<p class="text-text-muted">
						{multipleModesDescription || 'Enjoy various game modes including flag guessing, shape identification, capitals, and more.'}
					</p>
				</article>
				<article class="p-4">
					<div class="text-4xl mb-3" aria-hidden="true">🌏</div>
					<h3 class="text-xl font-semibold text-sandy-light mb-2">{regionalFocusTitle || 'Regional Focus'}</h3>
					<p class="text-text-muted">
						{regionalFocusDescription || 'Choose specific regions to focus your learning on different parts of the world.'}
					</p>
				</article>
				<article class="p-4">
					<div class="text-4xl mb-3" aria-hidden="true">📊</div>
					<h3 class="text-xl font-semibold text-sandy-light mb-2">{trackProgressTitle || 'Track Progress'}</h3>
					<p class="text-text-muted">
						{trackProgressDescription || 'Monitor your learning progress and improve your knowledge of world geography.'}
					</p>
				</article>
			</div>
		</section>
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

	.world-map-overlay {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		z-index: 1;
		background: radial-gradient(ellipse 85% 85% at 50% 50%, rgba(236, 72, 153, 0.04) 0%, rgba(249, 168, 212, 0.05) 20%, rgba(167, 139, 250, 0.04) 40%, rgba(196, 181, 253, 0.06) 60%, transparent 80%, rgba(10, 14, 39, 0.3) 100%);
	}
	
	.world-map-svg {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%) scale(1.3);
		width: 115%;
		max-width: 2100px;
		height: auto;
		opacity: 0.052;
		mix-blend-mode: soft-light;
		filter: blur(5px) brightness(2) contrast(1.2) saturate(1.6);
		transition: opacity 0.3s ease, transform 0.3s ease, filter 0.3s ease;
	}
	
	/* Ensure content is above background */
	.home-page > :global(.max-w-6xl) {
		position: relative;
		z-index: 1;
	}
	
	/* Subtle animation on hover for interactivity */
	@media (hover: hover) {
		.home-page:hover .world-map-svg {
			opacity: 0.065;
		}
	}

	/* Light mode adjustments for better visibility */
	:global(:root.light) .world-map-svg {
		opacity: 0.18;
		mix-blend-mode: multiply;
		filter: blur(4px) brightness(0.6) contrast(1.4) saturate(1.2);
	}

	:global(:root.light) .world-map-overlay {
		background: radial-gradient(ellipse 85% 85% at 50% 50%, rgba(236, 72, 153, 0.02) 0%, rgba(249, 168, 212, 0.03) 20%, rgba(167, 139, 250, 0.02) 40%, rgba(196, 181, 253, 0.03) 60%, transparent 80%, rgba(255, 255, 255, 0.1) 100%);
	}

	@media (hover: hover) {
		:global(:root.light) .home-page:hover .world-map-svg {
			opacity: 0.22;
		}
	}
</style>
