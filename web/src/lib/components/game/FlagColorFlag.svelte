<script lang="ts">
	import { browser } from '$app/environment';
	import { onMount, tick } from 'svelte';

	/** Absolute URL path e.g. /assets/twemoji_flags_cca2/FR.svg */
	export let flagUrl: string;
	export let guessableId: string;
	/** Stable id per question (kept for callers; reserved for future use). */
	export let questionId: string;
	/** When true, show the original flag colors (correct answer reveal). */
	export let revealFullColor = false;
	/** Current guess — painted live into the hidden region while sliders move. */
	export let previewR = 128;
	export let previewG = 128;
	export let previewB = 128;

	let container: HTMLDivElement | null = null;
	let rawSvgText = '';
	let loadKey = '';

	async function loadSvg(): Promise<void> {
		const key = `${flagUrl}|${guessableId}|${questionId}`;
		if (!flagUrl || !guessableId || !questionId) {
			rawSvgText = '';
			paint();
			return;
		}
		if (key === loadKey && rawSvgText !== '') {
			paint();
			return;
		}
		loadKey = key;
		rawSvgText = '';
		paint();
		try {
			const res = await fetch(flagUrl);
			rawSvgText = await res.text();
		} catch {
			rawSvgText = '';
		}
		await tick();
		paint();
	}

	function paint(): void {
		if (!browser || !container) return;
		if (!rawSvgText) {
			container.innerHTML = '';
			return;
		}
		const doc = new DOMParser().parseFromString(rawSvgText, 'image/svg+xml');
		const svg = doc.documentElement;
		if (!svg || svg.nodeName.toLowerCase() !== 'svg') {
			container.innerHTML = '';
			return;
		}
		const safeId = guessableId.replace(/"/g, '');
		const target = svg.querySelector(`[data-fi-guess="${safeId}"]`);
		if (!revealFullColor && target instanceof SVGElement) {
			target.setAttribute('fill', `rgb(${previewR},${previewG},${previewB})`);
		}
		const imported = document.importNode(svg, true);
		container.innerHTML = '';
		container.appendChild(imported);
	}

	$: flagUrl, guessableId, questionId, void loadSvg();

	$: previewR, previewG, previewB, revealFullColor, rawSvgText, guessableId, container, paint();

	onMount(() => {
		void loadSvg();
	});
</script>

<div
	bind:this={container}
	class="w-full max-w-[280px] md:max-w-xs mx-auto [&_svg]:w-full [&_svg]:h-auto [&_svg]:rounded-[20px] [&_svg]:shadow-lg border border-white/10 bg-black/20 min-h-[120px]"
	aria-hidden={false}
></div>
