<script lang="ts">
	import { browser } from '$app/environment';
	import { onDestroy, tick } from 'svelte';
	import { getCachedFlagSvgText, setCachedFlagSvgText } from '$lib/utils/flagColorSvgCache';

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
	let fetchAbort: AbortController | null = null;

	function cancelFetch(): void {
		fetchAbort?.abort();
		fetchAbort = null;
	}

	onDestroy(() => cancelFetch());

	async function loadSvg(): Promise<void> {
		const key = `${flagUrl}|${guessableId}|${questionId}`;
		if (!flagUrl || !guessableId || !questionId) {
			cancelFetch();
			loadKey = '';
			rawSvgText = '';
			paint();
			return;
		}
		if (key === loadKey && rawSvgText !== '') {
			paint();
			return;
		}
		// Reactive + bind:this can fire twice on mount; do not abort an in-flight load for the same key.
		if (key === loadKey && fetchAbort !== null) {
			return;
		}

		const cached = getCachedFlagSvgText(flagUrl);
		if (cached) {
			loadKey = key;
			rawSvgText = cached;
			await tick();
			paint();
			return;
		}

		cancelFetch();
		const controller = new AbortController();
		fetchAbort = controller;
		const previousKey = loadKey;
		loadKey = key;
		if (previousKey !== key) {
			rawSvgText = '';
			paint();
		}
		try {
			const res = await fetch(flagUrl, { signal: controller.signal });
			if (controller.signal.aborted || key !== loadKey) return;
			const text = await res.text();
			if (controller.signal.aborted || key !== loadKey) return;
			setCachedFlagSvgText(flagUrl, text);
			rawSvgText = text;
		} catch {
			if (controller.signal.aborted || key !== loadKey) return;
			rawSvgText = '';
		} finally {
			if (fetchAbort === controller) {
				fetchAbort = null;
			}
		}
		if (key !== loadKey) return;
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
		const targets = svg.querySelectorAll(`[data-fi-guess="${safeId}"]`);
		if (!revealFullColor) {
			targets.forEach((el) => {
				if (el instanceof SVGElement) {
					el.setAttribute('fill', `rgb(${previewR},${previewG},${previewB})`);
				}
			});
		}
		const imported = document.importNode(svg, true);
		container.innerHTML = '';
		container.appendChild(imported);
	}

	$: flagUrl, guessableId, questionId, void loadSvg();

	$: previewR, previewG, previewB, revealFullColor, rawSvgText, guessableId, container, paint();
</script>

<div
	bind:this={container}
	class="w-full max-w-[280px] md:max-w-xs mx-auto lg:mx-0 leading-none [&_svg]:block [&_svg]:w-full [&_svg]:h-auto [&_svg]:rounded-[20px] min-h-[120px]"
	aria-hidden={false}
></div>
