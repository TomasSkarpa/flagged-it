<script lang="ts">
	import { hsvToRgb } from '$lib/utils/colorHsv';

	export let hue = 200;
	export let satPct = 85;
	export let valPct = 75;

	export let disabled = false;
	/** Accessible name for the hue slider (e.g. translated). */
	export let hueAriaLabel = 'Hue';
	/** `game` uses a larger plane and hue strip for the flag color round layout. */
	export let size: 'default' | 'game' = 'default';

	let planeEl: HTMLDivElement | null = null;
	let dragging = false;

	$: hueN = Math.min(359, Math.max(0, Math.round(Number(hue))));
	$: chroma = hsvToRgb(hueN, 1, 1);
	$: planeBg = `rgb(${chroma.r},${chroma.g},${chroma.b})`;

	$: hueTrackMaxClass = size === 'game' ? 'max-w-[260px]' : 'max-w-[200px]';
	$: rootClass =
		size === 'game'
			? 'w-full max-w-[min(100%,300px)] min-w-0'
			: 'w-full max-w-[220px] min-w-0';
	$: planeMaxH = size === 'game' ? 'max-h-[min(48vw,220px)] sm:max-h-[240px]' : 'max-h-[200px]';

	function pickAtPointer(ev: PointerEvent, el: HTMLElement): void {
		const rect = el.getBoundingClientRect();
		const x = Math.min(Math.max(0, ev.clientX - rect.left), rect.width);
		const y = Math.min(Math.max(0, ev.clientY - rect.top), rect.height);
		satPct = Math.round((x / rect.width) * 100);
		valPct = Math.round((1 - y / rect.height) * 100);
	}

	function onPlanePointerDown(ev: PointerEvent): void {
		if (disabled || !planeEl) return;
		dragging = true;
		planeEl.setPointerCapture(ev.pointerId);
		pickAtPointer(ev, planeEl);
	}

	function onPlanePointerMove(ev: PointerEvent): void {
		if (!dragging || disabled || !planeEl) return;
		pickAtPointer(ev, planeEl);
	}

	function endPlaneDrag(ev: PointerEvent): void {
		if (!dragging || !planeEl) return;
		dragging = false;
		try {
			planeEl.releasePointerCapture(ev.pointerId);
		} catch {
			/* ignore */
		}
	}
</script>

<svelte:window on:pointerup={endPlaneDrag} on:pointercancel={endPlaneDrag} />

<div
	class="flag-color-svplane flex flex-col items-center gap-4 touch-manipulation select-none {rootClass} mx-auto"
	aria-label="Saturation and brightness plane with hue"
>
	<div class="relative w-full aspect-[4/3] {planeMaxH} rounded-2xl border border-white/15 shadow-inner overflow-hidden">
		<div
			bind:this={planeEl}
			class="absolute inset-0 cursor-crosshair touch-none {disabled ? 'opacity-50 pointer-events-none' : ''}"
			style="background-color: {planeBg}; background-image: linear-gradient(to top, rgb(0,0,0), transparent), linear-gradient(to right, rgb(255,255,255), transparent);"
			on:pointerdown={onPlanePointerDown}
			on:pointermove={onPlanePointerMove}
			role="presentation"
		></div>
		<div
			class="pointer-events-none absolute w-4 h-4 rounded-full border-2 border-white shadow-md ring-1 ring-black/60 -translate-x-1/2 -translate-y-1/2"
			style="left: {satPct}%; top: {100 - valPct}%;"
			aria-hidden="true"
		></div>
	</div>

	<div class="flex flex-col gap-2 w-full px-1">
		<div
			class="relative flex items-center rounded-full border border-white/15 shadow-inner bg-neutral-900 overflow-hidden mx-auto w-full {hueTrackMaxClass}"
			style="height: 44px;"
		>
			<div
				class="pointer-events-none absolute inset-1 rounded-full"
				style="background: linear-gradient(to right, #f00 0%, #ff0 17%, #0f0 33%, #0ff 50%, #00f 67%, #f0f 83%, #f00 100%);"
			></div>
			<input
				type="range"
				min="0"
				max="359"
				value={hueN}
				{disabled}
				class="absolute z-10 cursor-pointer opacity-0 flag-color-svplane-hue inset-x-0 w-full"
				style="height: 44px; top: 0;"
				aria-label={hueAriaLabel}
				on:input={(e) => {
					hue = +e.currentTarget.value;
				}}
			/>
			<div
				class="pointer-events-none absolute top-1/2 -translate-y-1/2 h-[85%] w-1 bg-white rounded-full shadow-md ring-1 ring-black/50 z-[5]"
				style="left: calc(8px + (100% - 16px) * {hueN / 359});"
			></div>
		</div>
		<span class="text-[10px] uppercase tracking-wider text-white/75 font-semibold text-center">{hueAriaLabel}</span>
	</div>
</div>

<style>
	:global(.flag-color-svplane-hue) {
		-webkit-appearance: none;
		appearance: none;
		background: transparent;
	}
	:global(.flag-color-svplane-hue::-webkit-slider-thumb) {
		-webkit-appearance: none;
		appearance: none;
		width: 28px;
		height: 28px;
		opacity: 0;
	}
	:global(.flag-color-svplane-hue::-moz-range-thumb) {
		width: 28px;
		height: 28px;
		opacity: 0;
		border: none;
	}
</style>
