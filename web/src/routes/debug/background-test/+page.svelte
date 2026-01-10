<script lang="ts">
	import { onMount } from 'svelte';
	import GameCard from '$lib/components/ui/GameCard.svelte';

	let currentIndex = 0;
	let previewMode = false;

	// 10 different background style variations
	const backgroundStyles = [
		{
			name: 'Screen Blend - Subtle',
			description: 'Screen blend mode with very low opacity, blurred edges',
			overlay: 'radial-gradient(ellipse 90% 90% at 50% 50%, transparent 0%, transparent 50%, rgba(10, 14, 39, 0.3) 70%, rgba(10, 14, 39, 0.7) 100%)',
			svgStyle: {
				opacity: '0.04',
				mixBlendMode: 'screen',
				filter: 'blur(3px) brightness(2) contrast(0.8)',
				transform: 'translate(-50%, -50%) scale(1.15)',
				width: '100%',
				maxWidth: '1800px'
			}
		},
		{
			name: 'Multiply Burn Effect',
			description: 'Multiply blend with burn-like filter, stretched full width',
			overlay: 'radial-gradient(ellipse 100% 100% at 50% 50%, transparent 20%, rgba(10, 14, 39, 0.4) 60%, rgba(10, 14, 39, 0.9) 100%)',
			svgStyle: {
				opacity: '0.06',
				mixBlendMode: 'multiply',
				filter: 'blur(2px) brightness(0.7) contrast(1.3)',
				transform: 'translate(-50%, -50%) scale(1.3)',
				width: '100%',
				maxWidth: '2000px'
			}
		},
		{
			name: 'Overlay with Hard Light',
			description: 'Overlay blend with hard light filter, burned edges',
			overlay: 'radial-gradient(ellipse 95% 95% at 50% 50%, transparent 10%, rgba(10, 14, 39, 0.2) 40%, rgba(10, 14, 39, 0.8) 80%, rgba(10, 14, 39, 0.95) 100%)',
			svgStyle: {
				opacity: '0.08',
				mixBlendMode: 'overlay',
				filter: 'blur(1.5px) brightness(1.5) contrast(1.2) saturate(0.8)',
				transform: 'translate(-50%, -50%) scale(1.2)',
				width: '100%',
				maxWidth: '1900px'
			}
		},
		{
			name: 'Color Dodge - Bright Burn',
			description: 'Color dodge for bright burn effect, fully stretched',
			overlay: 'radial-gradient(ellipse 85% 85% at 50% 50%, transparent 0%, transparent 45%, rgba(10, 14, 39, 0.5) 70%, rgba(10, 14, 39, 0.95) 100%)',
			svgStyle: {
				opacity: '0.05',
				mixBlendMode: 'color-dodge',
				filter: 'blur(4px) brightness(3) contrast(1.5)',
				transform: 'translate(-50%, -50%) scale(1.4)',
				width: '120%',
				maxWidth: '2200px'
			}
		},
		{
			name: 'Soft Light - Gentle',
			description: 'Soft light blend with gentle burn, vignette edges',
			overlay: 'radial-gradient(ellipse 80% 80% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.1) 50%, rgba(10, 14, 39, 0.6) 80%, rgba(10, 14, 39, 0.9) 100%)',
			svgStyle: {
				opacity: '0.07',
				mixBlendMode: 'soft-light',
				filter: 'blur(2.5px) brightness(1.8) contrast(1.1)',
				transform: 'translate(-50%, -50%) scale(1.25)',
				width: '110%',
				maxWidth: '2000px'
			}
		},
		{
			name: 'Darken with Strong Burn',
			description: 'Darken blend mode, strong burn effect at edges',
			overlay: 'radial-gradient(ellipse 75% 75% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.2) 40%, rgba(10, 14, 39, 0.7) 65%, rgba(10, 14, 39, 1) 100%)',
			svgStyle: {
				opacity: '0.1',
				mixBlendMode: 'darken',
				filter: 'blur(3px) brightness(0.5) contrast(1.8)',
				transform: 'translate(-50%, -50%) scale(1.35)',
				width: '115%',
				maxWidth: '2100px'
			}
		},
		{
			name: 'Difference - High Contrast',
			description: 'Difference blend for high contrast, burned into background',
			overlay: 'radial-gradient(ellipse 90% 90% at 50% 50%, transparent 0%, transparent 55%, rgba(10, 14, 39, 0.4) 75%, rgba(10, 14, 39, 0.85) 100%)',
			svgStyle: {
				opacity: '0.03',
				mixBlendMode: 'difference',
				filter: 'blur(2px) brightness(2.5) contrast(2) invert(0.1)',
				transform: 'translate(-50%, -50%) scale(1.3)',
				width: '100%',
				maxWidth: '2000px'
			}
		},
		{
			name: 'Exclusion - Subtle Burn',
			description: 'Exclusion blend, subtle burn with soft edges',
			overlay: 'radial-gradient(ellipse 88% 88% at 50% 50%, transparent 0%, transparent 48%, rgba(10, 14, 39, 0.35) 72%, rgba(10, 14, 39, 0.8) 100%)',
			svgStyle: {
				opacity: '0.045',
				mixBlendMode: 'exclusion',
				filter: 'blur(2.5px) brightness(1.6) contrast(1.4)',
				transform: 'translate(-50%, -50%) scale(1.28)',
				width: '105%',
				maxWidth: '1950px'
			}
		},
		{
			name: 'Hard Light - Sharp Burn',
			description: 'Hard light blend, sharp burn edges, fully stretched',
			overlay: 'radial-gradient(ellipse 92% 92% at 50% 50%, transparent 0%, transparent 52%, rgba(10, 14, 39, 0.45) 74%, rgba(10, 14, 39, 0.88) 100%)',
			svgStyle: {
				opacity: '0.055',
				mixBlendMode: 'hard-light',
				filter: 'blur(1px) brightness(1.4) contrast(1.6) saturate(0.9)',
				transform: 'translate(-50%, -50%) scale(1.32)',
				width: '125%',
				maxWidth: '2150px'
			}
		},
		{
			name: 'Color Burn - Intense',
			description: 'Color burn blend, intense burned effect, maximum stretch',
			overlay: 'radial-gradient(ellipse 70% 70% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.25) 35%, rgba(10, 14, 39, 0.75) 60%, rgba(10, 14, 39, 1) 100%)',
			svgStyle: {
				opacity: '0.09',
				mixBlendMode: 'color-burn',
				filter: 'blur(3.5px) brightness(0.6) contrast(2) saturate(1.2)',
				transform: 'translate(-50%, -50%) scale(1.45)',
				width: '130%',
				maxWidth: '2400px'
			}
		},
		{
			name: 'Cyberpunk Neon Glow',
			description: 'Vibrant neon outline with cyberpunk aesthetic, electric edges',
			overlay: 'radial-gradient(ellipse 85% 85% at 50% 50%, rgba(99, 102, 241, 0.15) 0%, rgba(139, 92, 246, 0.1) 30%, transparent 60%, rgba(99, 102, 241, 0.08) 85%, transparent 100%)',
			svgStyle: {
				opacity: '0.12',
				mixBlendMode: 'screen',
				filter: 'blur(0.5px) brightness(4) contrast(2) saturate(2) hue-rotate(240deg)',
				transform: 'translate(-50%, -50%) scale(1.2)',
				width: '110%',
				maxWidth: '1900px'
			}
		},
		{
			name: 'Glitch Distortion',
			description: 'Digital glitch effect with chromatic aberration and displacement',
			overlay: 'radial-gradient(ellipse 90% 90% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.2) 50%, rgba(10, 14, 39, 0.6) 80%, rgba(10, 14, 39, 0.9) 100%)',
			svgStyle: {
				opacity: '0.15',
				mixBlendMode: 'difference',
				filter: 'blur(1px) brightness(1.8) contrast(3) saturate(1.5) hue-rotate(90deg)',
				transform: 'translate(-50%, -50%) scale(1.25) translateX(2px)',
				width: '105%',
				maxWidth: '2000px'
			}
		},
		{
			name: 'Halftone Dots',
			description: 'Retro halftone newspaper print effect with dot pattern',
			overlay: 'radial-gradient(ellipse 80% 80% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.15) 45%, rgba(10, 14, 39, 0.5) 75%, rgba(10, 14, 39, 0.85) 100%), repeating-conic-gradient(from 0deg at 50% 50%, transparent 0%, rgba(255, 255, 255, 0.02) 1%, transparent 2%)',
			svgStyle: {
				opacity: '0.1',
				mixBlendMode: 'multiply',
				filter: 'blur(2px) brightness(0.8) contrast(2.5) saturate(0.5)',
				transform: 'translate(-50%, -50%) scale(1.3)',
				width: '115%',
				maxWidth: '2100px'
			}
		},
		{
			name: 'CRT Scanlines',
			description: 'Retro CRT monitor effect with horizontal scanlines and flicker',
			overlay: 'radial-gradient(ellipse 90% 90% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.3) 60%, rgba(10, 14, 39, 0.8) 100%), repeating-linear-gradient(0deg, transparent 0px, transparent 1px, rgba(0, 255, 0, 0.03) 1px, rgba(0, 255, 0, 0.03) 2px)',
			svgStyle: {
				opacity: '0.08',
				mixBlendMode: 'screen',
				filter: 'blur(1.5px) brightness(2) contrast(1.8) saturate(1.2)',
				transform: 'translate(-50%, -50%) scale(1.28)',
				width: '108%',
				maxWidth: '1950px'
			}
		},
		{
			name: 'Watercolor Bleed',
			description: 'Organic watercolor effect with soft bleeding edges and color wash',
			overlay: 'radial-gradient(ellipse 75% 75% at 50% 50%, rgba(6, 182, 212, 0.08) 0%, transparent 30%, rgba(99, 102, 241, 0.12) 60%, transparent 85%, rgba(139, 92, 246, 0.1) 100%)',
			svgStyle: {
				opacity: '0.11',
				mixBlendMode: 'soft-light',
				filter: 'blur(4px) brightness(1.6) contrast(1.4) saturate(1.8)',
				transform: 'translate(-50%, -50%) scale(1.35)',
				width: '120%',
				maxWidth: '2200px'
			}
		},
		{
			name: 'Voxel Pixel Art',
			description: 'Pixelated voxel effect with sharp edges and blocky aesthetic',
			overlay: 'radial-gradient(ellipse 88% 88% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.25) 50%, rgba(10, 14, 39, 0.7) 80%, rgba(10, 14, 39, 0.95) 100%), repeating-conic-gradient(from 0deg at 50% 50%, transparent 0deg, rgba(255, 255, 255, 0.01) 15deg, transparent 30deg)',
			svgStyle: {
				opacity: '0.13',
				mixBlendMode: 'hard-light',
				filter: 'contrast(2.2) brightness(1.3) saturate(1.1)',
				transform: 'translate(-50%, -50%) scale(1.22)',
				width: '102%',
				maxWidth: '1850px'
			}
		},
		{
			name: 'Double Exposure Ghost',
			description: 'Ethereal double exposure with ghostly overlay and transparency',
			overlay: 'radial-gradient(ellipse 95% 95% at 50% 50%, transparent 0%, rgba(6, 182, 212, 0.05) 35%, rgba(99, 102, 241, 0.08) 65%, transparent 100%)',
			svgStyle: {
				opacity: '0.06',
				mixBlendMode: 'exclusion',
				filter: 'blur(3px) brightness(2.5) contrast(1.6) saturate(1.3) opacity(0.7)',
				transform: 'translate(-50%, -50%) scale(1.18) translateX(-1%) translateY(1%)',
				width: '112%',
				maxWidth: '2050px'
			}
		},
		{
			name: 'Graffiti Stencil',
			description: 'Urban graffiti stencil effect with hard edges and spray paint texture',
			overlay: 'radial-gradient(ellipse 82% 82% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.2) 40%, rgba(10, 14, 39, 0.6) 70%, rgba(10, 14, 39, 1) 100%)',
			svgStyle: {
				opacity: '0.16',
				mixBlendMode: 'multiply',
				filter: 'blur(0px) brightness(0.4) contrast(3) saturate(1.4)',
				transform: 'translate(-50%, -50%) scale(1.4)',
				width: '125%',
				maxWidth: '2300px'
			}
		},
		{
			name: 'Warped Perspective',
			description: 'Distorted warped perspective with fisheye-like distortion effect',
			overlay: 'radial-gradient(ellipse 70% 90% at 50% 45%, transparent 0%, rgba(10, 14, 39, 0.18) 45%, rgba(10, 14, 39, 0.65) 75%, rgba(10, 14, 39, 0.92) 100%)',
			svgStyle: {
				opacity: '0.14',
				mixBlendMode: 'overlay',
				filter: 'blur(2px) brightness(1.4) contrast(2) saturate(1.6)',
				transform: 'translate(-50%, -50%) scale(1.38) scaleY(1.15) rotate(-1deg)',
				width: '118%',
				maxWidth: '2150px'
			}
		},
		{
			name: 'Holographic Rainbow',
			description: 'Iridescent holographic effect with rainbow color shifts and prismatic edges',
			overlay: 'radial-gradient(ellipse 90% 90% at 50% 50%, rgba(99, 102, 241, 0.06) 0%, rgba(139, 92, 246, 0.04) 25%, rgba(236, 72, 153, 0.05) 50%, rgba(6, 182, 212, 0.06) 75%, transparent 100%)',
			svgStyle: {
				opacity: '0.1',
				mixBlendMode: 'color-dodge',
				filter: 'blur(2px) brightness(3) contrast(1.8) saturate(2.5) hue-rotate(45deg)',
				transform: 'translate(-50%, -50%) scale(1.26)',
				width: '114%',
				maxWidth: '2080px'
			}
		},
		{
			name: 'X-Ray Negative',
			description: 'X-ray negative inversion with high contrast medical imaging aesthetic',
			overlay: 'radial-gradient(ellipse 85% 85% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.3) 55%, rgba(10, 14, 39, 0.75) 80%, rgba(10, 14, 39, 0.98) 100%)',
			svgStyle: {
				opacity: '0.12',
				mixBlendMode: 'difference',
				filter: 'blur(1px) brightness(0.3) contrast(3) saturate(0) invert(1)',
				transform: 'translate(-50%, -50%) scale(1.3)',
				width: '109%',
				maxWidth: '1980px'
			}
		},
		{
			name: 'Oil Painting Brush',
			description: 'Impressionist oil painting with visible brush strokes and texture',
			overlay: 'radial-gradient(ellipse 78% 78% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.12) 42%, rgba(10, 14, 39, 0.55) 72%, rgba(10, 14, 39, 0.88) 100%)',
			svgStyle: {
				opacity: '0.095',
				mixBlendMode: 'soft-light',
				filter: 'blur(5px) brightness(1.5) contrast(1.6) saturate(1.7)',
				transform: 'translate(-50%, -50%) scale(1.32) rotate(0.5deg)',
				width: '116%',
				maxWidth: '2120px'
			}
		},
		{
			name: 'Punk Zine Cutout',
			description: 'DIY punk zine aesthetic with torn edges and high contrast photocopy texture',
			overlay: 'radial-gradient(ellipse 72% 72% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.28) 38%, rgba(10, 14, 39, 0.68) 65%, rgba(10, 14, 39, 1) 92%)',
			svgStyle: {
				opacity: '0.18',
				mixBlendMode: 'darken',
				filter: 'blur(0.5px) brightness(0.5) contrast(2.8) saturate(1.2)',
				transform: 'translate(-50%, -50%) scale(1.42) rotate(-0.8deg)',
				width: '128%',
				maxWidth: '2350px'
			}
		}
	];

	$: currentStyle = backgroundStyles[currentIndex];

	function nextStyle() {
		currentIndex = (currentIndex + 1) % backgroundStyles.length;
	}

	function previousStyle() {
		currentIndex = currentIndex === 0 ? backgroundStyles.length - 1 : currentIndex - 1;
	}

	function togglePreview() {
		previewMode = !previewMode;
	}

	// Keyboard navigation
	onMount(() => {
		function handleKeyPress(e: KeyboardEvent) {
			if (e.key === 'ArrowLeft') previousStyle();
			if (e.key === 'ArrowRight') nextStyle();
			if (e.key === ' ' || e.key === 'Enter') {
				e.preventDefault();
				togglePreview();
			}
		}
		window.addEventListener('keydown', handleKeyPress);
		return () => window.removeEventListener('keydown', handleKeyPress);
	});
</script>

<svelte:head>
	<title>Background Test - Flagged It</title>
	<meta name="description" content="Test different background styles for the world map silhouette" />
</svelte:head>

<div class="background-test-page min-h-screen p-4 md:p-8 relative overflow-hidden">
	<!-- World Map Background with current style -->
	<div class="world-map-background">
		<div class="world-map-overlay" style="background: {currentStyle.overlay};"></div>
		<img 
			src="/assets/world_map_silhouette.svg" 
			alt="" 
			class="world-map-svg"
			style="opacity: {currentStyle.svgStyle.opacity}; mix-blend-mode: {currentStyle.svgStyle.mixBlendMode}; filter: {currentStyle.svgStyle.filter}; transform: {currentStyle.svgStyle.transform}; width: {currentStyle.svgStyle.width}; max-width: {currentStyle.svgStyle.maxWidth};"
			aria-hidden="true"
		/>
	</div>

	<div class="max-w-6xl mx-auto relative z-10">
		<!-- Control Panel -->
		<GameCard class="mb-8 sticky top-20 z-20">
			<div class="space-y-4">
				<div class="flex items-center justify-between mb-4">
					<div>
						<h1 class="text-3xl font-bold text-sandy-light mb-2">
							Background Style Test
						</h1>
						<p class="text-text-muted">
							Style {currentIndex + 1} of {backgroundStyles.length}: {currentStyle.name}
						</p>
					</div>
					<button
						class="btn-secondary px-4 py-2 rounded-lg font-semibold transition-all"
						on:click={togglePreview}
					>
						{previewMode ? 'Show Controls' : 'Full Preview Mode'}
					</button>
				</div>

				{#if !previewMode}
					<div>
						<p class="text-sm text-text-muted mb-3">{currentStyle.description}</p>
						<div class="grid grid-cols-2 gap-4 text-xs text-text-dark font-mono bg-surface-light/50 p-3 rounded">
							<div>
								<p><strong>Opacity:</strong> {currentStyle.svgStyle.opacity}</p>
								<p><strong>Blend Mode:</strong> {currentStyle.svgStyle.mixBlendMode}</p>
								<p><strong>Transform:</strong> {currentStyle.svgStyle.transform.split(' ').slice(-1)[0]}</p>
							</div>
							<div>
								<p><strong>Filter:</strong> {currentStyle.svgStyle.filter.substring(0, 40)}...</p>
								<p><strong>Width:</strong> {currentStyle.svgStyle.width}</p>
								<p><strong>Max Width:</strong> {currentStyle.svgStyle.maxWidth}</p>
							</div>
						</div>
					</div>

					<div class="flex gap-4 items-center justify-center">
						<button
							class="btn-secondary px-4 py-2 rounded-lg font-semibold transition-all disabled:opacity-50 disabled:cursor-not-allowed"
							on:click={previousStyle}
							disabled={currentIndex === 0}
						>
							← Previous
						</button>
						<div class="flex gap-2">
							{#each backgroundStyles as _, i}
								<button
									class="w-3 h-3 rounded-full transition-all {i === currentIndex ? 'bg-primary w-8' : 'bg-white/20 hover:bg-white/30'}"
									on:click={() => { currentIndex = i; }}
									aria-label="Go to style {i + 1}"
								></button>
							{/each}
						</div>
						<button
							class="btn-primary px-4 py-2 rounded-lg font-semibold transition-all"
							on:click={nextStyle}
						>
							Next →
						</button>
					</div>

					<div class="text-center text-xs text-text-muted pt-2 border-t border-white/10">
						<p>Use ← → arrow keys to navigate, Space/Enter to toggle preview mode</p>
					</div>
				{/if}
			</div>
		</GameCard>

		<!-- Preview Content (similar to home page) -->
		{#if previewMode}
			<div class="space-y-8">
				<!-- Hero Section -->
				<div class="text-center mb-16">
					<h1 class="text-5xl md:text-7xl font-bold mb-4">
						<span class="mr-2 emoji-blue">🌍</span><span class="gradient-text">Flagged It</span>
					</h1>
					<p class="text-xl md:text-2xl text-text-muted mb-2">
						Test Your Geography Knowledge
					</p>
					<p class="text-lg text-text-dark">
						Play fun games and learn about countries around the world!
					</p>
				</div>

				<!-- Sample Game Cards -->
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-12">
					<GameCard class="text-left">
						<div class="flex items-start gap-4">
							<div class="text-5xl leading-none">🚩</div>
							<div class="flex-1">
								<h2 class="text-2xl font-bold text-sandy-light mb-2 leading-none mt-0">
									Guess by Flag
								</h2>
								<p class="text-text-muted mb-4">
									Identify countries by their flags
								</p>
								<span class="inline-block px-4 py-2 bg-primary/20 border border-primary rounded-lg text-primary font-semibold">
									Play Now →
								</span>
							</div>
						</div>
					</GameCard>

					<GameCard class="text-left">
						<div class="flex items-start gap-4">
							<div class="text-5xl leading-none">🗺️</div>
							<div class="flex-1">
								<h2 class="text-2xl font-bold text-sandy-light mb-2 leading-none mt-0">
									Guess by Shape
								</h2>
								<p class="text-text-muted mb-4">
									Identify countries by their shape
								</p>
								<span class="inline-block px-4 py-2 bg-primary/20 border border-primary rounded-lg text-primary font-semibold">
									Play Now →
								</span>
							</div>
						</div>
					</GameCard>
				</div>

				<!-- Floating Navigation -->
				<div class="fixed bottom-8 left-1/2 transform -translate-x-1/2 z-30">
					<GameCard class="p-4">
						<div class="flex gap-4 items-center">
							<button
								class="btn-secondary px-4 py-2 text-sm rounded-lg font-semibold transition-all"
								on:click={previousStyle}
							>
								← Prev
							</button>
							<span class="text-sm text-text-muted px-4">
								{currentIndex + 1} / {backgroundStyles.length}
							</span>
							<button
								class="btn-primary px-4 py-2 text-sm rounded-lg font-semibold transition-all"
								on:click={nextStyle}
							>
								Next →
							</button>
							<button
								class="btn-secondary px-4 py-2 text-sm rounded-lg font-semibold transition-all"
								on:click={togglePreview}
							>
								Exit Preview
							</button>
						</div>
					</GameCard>
				</div>
			</div>
		{:else}
			<!-- Style List -->
			<GameCard>
				<h2 class="text-2xl font-bold text-sandy-light mb-4">All Background Styles</h2>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					{#each backgroundStyles as style, i}
						<button
							class="text-left p-4 rounded-lg border-2 transition-all cursor-pointer {i === currentIndex ? 'border-primary bg-primary/10' : 'border-white/10 bg-surface-light/50 hover:border-primary/50'}"
							on:click={() => { currentIndex = i; }}
						>
							<div class="flex items-center gap-2 mb-2">
								<span class="text-xs font-bold text-primary">#{i + 1}</span>
								<h3 class="font-semibold text-text-light">{style.name}</h3>
							</div>
							<p class="text-sm text-text-muted">{style.description}</p>
						</button>
					{/each}
				</div>
			</GameCard>
		{/if}
	</div>
</div>

<style>
	.background-test-page {
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
	}

	.world-map-svg {
		position: absolute;
		top: 50%;
		left: 50%;
		height: auto;
		z-index: 0;
		transition: opacity 0.4s ease, transform 0.4s ease, filter 0.4s ease;
	}

	.background-test-page > :global(.max-w-6xl) {
		position: relative;
		z-index: 1;
	}
</style>
