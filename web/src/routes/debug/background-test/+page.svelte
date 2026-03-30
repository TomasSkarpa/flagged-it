<script lang="ts">
	import { onMount } from 'svelte';
	import GameCard from '$lib/components/ui/GameCard.svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';

	$: currentLocale = $locale;
	$: backgroundTestTitle = t('debug.background_test.document_title', undefined, currentLocale);
	$: backgroundTestDescription = t('debug.background_test.meta_description', undefined, currentLocale);

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
		},
		// Nature / Organic
		{
			name: 'Forest Mist',
			description: 'Dense fog over forest silhouette with soft green gradient overlay',
			overlay: 'radial-gradient(ellipse 90% 90% at 50% 50%, transparent 0%, rgba(16, 185, 129, 0.06) 35%, rgba(5, 150, 105, 0.1) 60%, transparent 90%, rgba(10, 14, 39, 0.4) 100%)',
			svgStyle: {
				opacity: '0.05',
				mixBlendMode: 'soft-light',
				filter: 'blur(4px) brightness(1.4) contrast(1.3) saturate(1.2) hue-rotate(120deg)',
				transform: 'translate(-50%, -50%) scale(1.2)',
				width: '110%',
				maxWidth: '2000px'
			}
		},
		{
			name: 'Mountain Sunset',
			description: 'Mountains with warm gradient glow and subtle lens blur',
			overlay: 'radial-gradient(ellipse 85% 85% at 50% 50%, transparent 0%, rgba(245, 158, 11, 0.08) 30%, rgba(217, 119, 6, 0.06) 55%, rgba(10, 14, 39, 0.3) 80%, rgba(10, 14, 39, 0.7) 100%)',
			svgStyle: {
				opacity: '0.055',
				mixBlendMode: 'overlay',
				filter: 'blur(3px) brightness(1.6) contrast(1.4) saturate(1.3)',
				transform: 'translate(-50%, -50%) scale(1.25)',
				width: '115%',
				maxWidth: '2100px'
			}
		},
		{
			name: 'Ocean Waves',
			description: 'Rolling waves with soft blue ripple overlay and slight motion blur',
			overlay: 'radial-gradient(ellipse 95% 95% at 50% 50%, rgba(6, 182, 212, 0.05) 0%, transparent 30%, rgba(59, 130, 246, 0.08) 60%, transparent 85%, rgba(10, 14, 39, 0.35) 100%)',
			svgStyle: {
				opacity: '0.06',
				mixBlendMode: 'screen',
				filter: 'blur(3.5px) brightness(2) contrast(1.2) saturate(1.5) hue-rotate(180deg)',
				transform: 'translate(-50%, -50%) scale(1.22)',
				width: '108%',
				maxWidth: '1950px'
			}
		},
		{
			name: 'Desert Dunes',
			description: 'Sand dunes with warm orange overlay and subtle grain texture',
			overlay: 'radial-gradient(ellipse 88% 88% at 50% 50%, transparent 0%, rgba(249, 115, 22, 0.07) 40%, rgba(217, 119, 6, 0.05) 65%, rgba(10, 14, 39, 0.4) 90%, rgba(10, 14, 39, 0.75) 100%), repeating-conic-gradient(from 0deg at 50% 50%, transparent 0%, rgba(255, 255, 255, 0.01) 1%, transparent 3%)',
			svgStyle: {
				opacity: '0.065',
				mixBlendMode: 'multiply',
				filter: 'blur(2.5px) brightness(1.3) contrast(1.5) saturate(1.4)',
				transform: 'translate(-50%, -50%) scale(1.28)',
				width: '112%',
				maxWidth: '2050px'
			}
		},
		// Space / Sci-Fi
		{
			name: 'Galaxy Nebula',
			description: 'Colorful star clouds with faint glows and star sparkle overlay',
			overlay: 'radial-gradient(ellipse 92% 92% at 50% 50%, rgba(139, 92, 246, 0.04) 0%, rgba(99, 102, 241, 0.06) 25%, rgba(236, 72, 153, 0.05) 50%, rgba(6, 182, 212, 0.04) 75%, transparent 100%), repeating-conic-gradient(from 0deg at 50% 50%, transparent 0deg, rgba(255, 255, 255, 0.015) 10deg, transparent 20deg)',
			svgStyle: {
				opacity: '0.05',
				mixBlendMode: 'color-dodge',
				filter: 'blur(3px) brightness(3.5) contrast(1.6) saturate(2)',
				transform: 'translate(-50%, -50%) scale(1.24)',
				width: '114%',
				maxWidth: '2080px'
			}
		},
		{
			name: 'Planet Rings',
			description: 'Planet with rings silhouette and radial glow around it',
			overlay: 'radial-gradient(ellipse 75% 75% at 50% 50%, transparent 0%, rgba(6, 182, 212, 0.06) 35%, rgba(99, 102, 241, 0.05) 55%, transparent 75%, rgba(10, 14, 39, 0.3) 95%, rgba(10, 14, 39, 0.7) 100%)',
			svgStyle: {
				opacity: '0.06',
				mixBlendMode: 'screen',
				filter: 'blur(4px) brightness(2.5) contrast(1.4) saturate(1.8)',
				transform: 'translate(-50%, -50%) scale(1.3)',
				width: '116%',
				maxWidth: '2120px'
			}
		},
		{
			name: 'Cyber Grid',
			description: 'Futuristic cityscape with neon gridlines and digital glitch overlay',
			overlay: 'radial-gradient(ellipse 90% 90% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.25) 50%, rgba(10, 14, 39, 0.6) 80%, rgba(10, 14, 39, 0.9) 100%), repeating-linear-gradient(0deg, transparent 0px, transparent 19px, rgba(99, 102, 241, 0.02) 19px, rgba(99, 102, 241, 0.02) 20px), repeating-linear-gradient(90deg, transparent 0px, transparent 19px, rgba(139, 92, 246, 0.02) 19px, rgba(139, 92, 246, 0.02) 20px)',
			svgStyle: {
				opacity: '0.07',
				mixBlendMode: 'screen',
				filter: 'blur(1px) brightness(2.2) contrast(1.8) saturate(1.6) hue-rotate(250deg)',
				transform: 'translate(-50%, -50%) scale(1.26) rotate(0.5deg)',
				width: '118%',
				maxWidth: '2150px'
			}
		},
		{
			name: 'Black Hole',
			description: 'Swirling dark center with light warp distortions',
			overlay: 'radial-gradient(ellipse 60% 60% at 50% 50%, rgba(10, 14, 39, 0.4) 0%, rgba(10, 14, 39, 0.25) 20%, transparent 40%, rgba(99, 102, 241, 0.05) 60%, rgba(236, 72, 153, 0.04) 80%, transparent 100%)',
			svgStyle: {
				opacity: '0.055',
				mixBlendMode: 'exclusion',
				filter: 'blur(5px) brightness(0.8) contrast(2.2) saturate(1.7)',
				transform: 'translate(-50%, -50%) scale(1.35) rotate(-1deg)',
				width: '120%',
				maxWidth: '2200px'
			}
		},
		// Retro / Digital
		{
			name: 'CRT Monitor',
			description: 'Horizontal scanlines with slight flicker and green/amber glow',
			overlay: 'radial-gradient(ellipse 88% 88% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.2) 50%, rgba(10, 14, 39, 0.6) 80%, rgba(10, 14, 39, 0.9) 100%), repeating-linear-gradient(0deg, transparent 0px, transparent 1px, rgba(0, 255, 0, 0.04) 1px, rgba(0, 255, 0, 0.04) 2px, transparent 2px, transparent 3px)',
			svgStyle: {
				opacity: '0.065',
				mixBlendMode: 'screen',
				filter: 'blur(2px) brightness(2.2) contrast(1.9) saturate(1.3) hue-rotate(90deg)',
				transform: 'translate(-50%, -50%) scale(1.25)',
				width: '110%',
				maxWidth: '2000px'
			}
		},
		{
			name: 'Pixel Art',
			description: 'Pixelated background with blocky shapes and low-res aesthetic',
			overlay: 'radial-gradient(ellipse 85% 85% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.22) 48%, rgba(10, 14, 39, 0.65) 78%, rgba(10, 14, 39, 0.92) 100%), repeating-conic-gradient(from 0deg at 50% 50%, transparent 0deg, rgba(255, 255, 255, 0.01) 5deg, transparent 10deg)',
			svgStyle: {
				opacity: '0.08',
				mixBlendMode: 'hard-light',
				filter: 'contrast(2.5) brightness(1.4) saturate(1.2) blur(0.5px)',
				transform: 'translate(-50%, -50%) scale(1.22)',
				width: '102%',
				maxWidth: '1850px'
			}
		},
		{
			name: 'Glitch Effect',
			description: 'Digital distortion with chromatic aberration and RGB shifts',
			overlay: 'radial-gradient(ellipse 92% 92% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.2) 52%, rgba(10, 14, 39, 0.6) 78%, rgba(10, 14, 39, 0.88) 100%)',
			svgStyle: {
				opacity: '0.07',
				mixBlendMode: 'difference',
				filter: 'blur(1.5px) brightness(1.9) contrast(2.8) saturate(1.7) hue-rotate(120deg)',
				transform: 'translate(-50%, -50%) scale(1.28) translateX(1px)',
				width: '107%',
				maxWidth: '1950px'
			}
		},
		{
			name: 'Halftone Print',
			description: 'Dot pattern with faded ink look and old newspaper feel',
			overlay: 'radial-gradient(ellipse 82% 82% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.18) 42%, rgba(10, 14, 39, 0.58) 72%, rgba(10, 14, 39, 0.85) 100%), repeating-radial-gradient(circle at 50% 50%, transparent 0px, transparent 3px, rgba(255, 255, 255, 0.02) 3px, rgba(255, 255, 255, 0.02) 4px, transparent 4px, transparent 8px)',
			svgStyle: {
				opacity: '0.075',
				mixBlendMode: 'multiply',
				filter: 'blur(2px) brightness(0.7) contrast(2.3) saturate(0.6)',
				transform: 'translate(-50%, -50%) scale(1.3)',
				width: '113%',
				maxWidth: '2060px'
			}
		},
		// Artistic / Abstract
		{
			name: 'Watercolor Wash',
			description: 'Organic bleeding colors with soft gradients and painterly edges',
			overlay: 'radial-gradient(ellipse 78% 78% at 50% 50%, rgba(6, 182, 212, 0.06) 0%, rgba(99, 102, 241, 0.08) 25%, transparent 45%, rgba(236, 72, 153, 0.07) 65%, transparent 85%, rgba(139, 92, 246, 0.09) 100%)',
			svgStyle: {
				opacity: '0.055',
				mixBlendMode: 'soft-light',
				filter: 'blur(5px) brightness(1.7) contrast(1.5) saturate(2)',
				transform: 'translate(-50%, -50%) scale(1.32) rotate(0.3deg)',
				width: '117%',
				maxWidth: '2130px'
			}
		},
		{
			name: 'Oil Brush Texture',
			description: 'Brush strokes layered with subtle highlights and shadows',
			overlay: 'radial-gradient(ellipse 80% 80% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.12) 40%, rgba(10, 14, 39, 0.52) 70%, rgba(10, 14, 39, 0.86) 100%)',
			svgStyle: {
				opacity: '0.068',
				mixBlendMode: 'soft-light',
				filter: 'blur(6px) brightness(1.5) contrast(1.7) saturate(1.8)',
				transform: 'translate(-50%, -50%) scale(1.3) rotate(-0.4deg)',
				width: '115%',
				maxWidth: '2100px'
			}
		},
		{
			name: 'Psychedelic Swirl',
			description: 'Colorful vortex with high saturation and animated rotation possible',
			overlay: 'radial-gradient(ellipse 70% 70% at 50% 50%, rgba(139, 92, 246, 0.07) 0%, rgba(236, 72, 153, 0.06) 20%, rgba(245, 158, 11, 0.05) 40%, rgba(6, 182, 212, 0.07) 60%, rgba(99, 102, 241, 0.06) 80%, transparent 100%)',
			svgStyle: {
				opacity: '0.06',
				mixBlendMode: 'color-dodge',
				filter: 'blur(3px) brightness(3) contrast(1.9) saturate(3) hue-rotate(45deg)',
				transform: 'translate(-50%, -50%) scale(1.33) rotate(2deg)',
				width: '119%',
				maxWidth: '2170px'
			}
		},
		{
			name: 'Geometric Shapes',
			description: 'Abstract triangles/circles with layered semi-transparent overlays',
			overlay: 'radial-gradient(ellipse 90% 90% at 50% 50%, transparent 0%, rgba(99, 102, 241, 0.04) 35%, rgba(139, 92, 246, 0.05) 55%, rgba(6, 182, 212, 0.04) 75%, transparent 100%), repeating-conic-gradient(from 30deg at 50% 50%, transparent 0deg, rgba(255, 255, 255, 0.01) 15deg, transparent 30deg, rgba(99, 102, 241, 0.01) 45deg, transparent 60deg)',
			svgStyle: {
				opacity: '0.05',
				mixBlendMode: 'exclusion',
				filter: 'blur(2.5px) brightness(2) contrast(1.6) saturate(1.4)',
				transform: 'translate(-50%, -50%) scale(1.24) rotate(-0.5deg)',
				width: '111%',
				maxWidth: '2020px'
			}
		},
		// Urban / Modern
		{
			name: 'Graffiti Wall',
			description: 'Urban wall textures with spray paint overlay and vibrant colors',
			overlay: 'radial-gradient(ellipse 75% 75% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.2) 38%, rgba(10, 14, 39, 0.62) 68%, rgba(10, 14, 39, 0.95) 100%)',
			svgStyle: {
				opacity: '0.085',
				mixBlendMode: 'multiply',
				filter: 'blur(1px) brightness(0.5) contrast(2.6) saturate(1.5)',
				transform: 'translate(-50%, -50%) scale(1.38) rotate(-0.6deg)',
				width: '123%',
				maxWidth: '2240px'
			}
		},
		{
			name: 'Neon City',
			description: 'Night cityscape with neon reflections and screen blend glow',
			overlay: 'radial-gradient(ellipse 88% 88% at 50% 50%, transparent 0%, rgba(99, 102, 241, 0.08) 30%, rgba(139, 92, 246, 0.06) 50%, rgba(236, 72, 153, 0.07) 70%, transparent 90%, rgba(10, 14, 39, 0.35) 100%)',
			svgStyle: {
				opacity: '0.07',
				mixBlendMode: 'screen',
				filter: 'blur(2px) brightness(2.8) contrast(1.9) saturate(2.2) hue-rotate(250deg)',
				transform: 'translate(-50%, -50%) scale(1.26)',
				width: '116%',
				maxWidth: '2110px'
			}
		},
		{
			name: 'Concrete Texture',
			description: 'Rough concrete background with subtle burn or multiply effect',
			overlay: 'radial-gradient(ellipse 85% 85% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.18) 45%, rgba(10, 14, 39, 0.6) 75%, rgba(10, 14, 39, 0.88) 100%)',
			svgStyle: {
				opacity: '0.072',
				mixBlendMode: 'multiply',
				filter: 'blur(2.5px) brightness(0.7) contrast(2.2) saturate(0.9)',
				transform: 'translate(-50%, -50%) scale(1.3)',
				width: '114%',
				maxWidth: '2080px'
			}
		},
		{
			name: 'Subway Map',
			description: 'Schematic lines over dark textured background with glowing paths',
			overlay: 'radial-gradient(ellipse 90% 90% at 50% 50%, transparent 0%, rgba(10, 14, 39, 0.25) 50%, rgba(10, 14, 39, 0.65) 80%, rgba(10, 14, 39, 0.92) 100%), repeating-linear-gradient(45deg, transparent 0px, transparent 29px, rgba(6, 182, 212, 0.03) 29px, rgba(6, 182, 212, 0.03) 30px), repeating-linear-gradient(-45deg, transparent 0px, transparent 29px, rgba(99, 102, 241, 0.03) 29px, rgba(99, 102, 241, 0.03) 30px)',
			svgStyle: {
				opacity: '0.066',
				mixBlendMode: 'screen',
				filter: 'blur(1.5px) brightness(2.4) contrast(2) saturate(1.6)',
				transform: 'translate(-50%, -50%) scale(1.27)',
				width: '112%',
				maxWidth: '2040px'
			}
		},
		// Seasonal / Mood
		{
			name: 'Winter Snow',
			description: 'Snowy landscape with falling flakes overlay and soft light',
			overlay: 'radial-gradient(ellipse 95% 95% at 50% 50%, rgba(236, 254, 255, 0.04) 0%, transparent 30%, rgba(226, 232, 240, 0.06) 60%, transparent 85%, rgba(10, 14, 39, 0.3) 100%), repeating-conic-gradient(from 0deg at 50% 50%, transparent 0deg, rgba(255, 255, 255, 0.02) 5deg, transparent 10deg)',
			svgStyle: {
				opacity: '0.055',
				mixBlendMode: 'screen',
				filter: 'blur(4px) brightness(2.2) contrast(1.3) saturate(0.8)',
				transform: 'translate(-50%, -50%) scale(1.23)',
				width: '109%',
				maxWidth: '1980px'
			}
		},
		{
			name: 'Autumn Leaves',
			description: 'Moving leaves with warm orange/brown overlay',
			overlay: 'radial-gradient(ellipse 88% 88% at 50% 50%, transparent 0%, rgba(245, 158, 11, 0.06) 35%, rgba(217, 119, 6, 0.05) 55%, rgba(180, 83, 9, 0.07) 75%, transparent 90%, rgba(10, 14, 39, 0.4) 100%)',
			svgStyle: {
				opacity: '0.06',
				mixBlendMode: 'overlay',
				filter: 'blur(3px) brightness(1.5) contrast(1.5) saturate(1.4)',
				transform: 'translate(-50%, -50%) scale(1.28) rotate(0.2deg)',
				width: '113%',
				maxWidth: '2060px'
			}
		},
		{
			name: 'Rainy Night',
			description: 'Raindrops with wet asphalt reflections and cool bluish overlay',
			overlay: 'radial-gradient(ellipse 92% 92% at 50% 50%, rgba(59, 130, 246, 0.05) 0%, rgba(37, 99, 235, 0.06) 25%, transparent 50%, rgba(30, 64, 175, 0.07) 75%, transparent 90%, rgba(10, 14, 39, 0.35) 100%), repeating-linear-gradient(90deg, transparent 0px, transparent 4px, rgba(59, 130, 246, 0.02) 4px, rgba(59, 130, 246, 0.02) 5px)',
			svgStyle: {
				opacity: '0.058',
				mixBlendMode: 'screen',
				filter: 'blur(3.5px) brightness(2) contrast(1.4) saturate(1.3) hue-rotate(200deg)',
				transform: 'translate(-50%, -50%) scale(1.25)',
				width: '111%',
				maxWidth: '2020px'
			}
		},
		{
			name: 'Spring Bloom',
			description: 'Pastel flowers in soft blur with gentle light bloom',
			overlay: 'radial-gradient(ellipse 85% 85% at 50% 50%, rgba(236, 72, 153, 0.04) 0%, rgba(249, 168, 212, 0.05) 20%, rgba(167, 139, 250, 0.04) 40%, rgba(196, 181, 253, 0.06) 60%, transparent 80%, rgba(10, 14, 39, 0.3) 100%)',
			svgStyle: {
				opacity: '0.052',
				mixBlendMode: 'soft-light',
				filter: 'blur(5px) brightness(2) contrast(1.2) saturate(1.6)',
				transform: 'translate(-50%, -50%) scale(1.3)',
				width: '115%',
				maxWidth: '2100px'
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
	<title>{backgroundTestTitle}</title>
	<meta name="description" content={backgroundTestDescription} />
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
