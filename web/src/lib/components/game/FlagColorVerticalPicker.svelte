<script lang="ts">
	import { hsvToRgb } from '$lib/utils/colorHsv';

	export let hue = 200;
	export let satPct = 85;
	export let valPct = 75;

	export let disabled = false;

	$: chroma = hsvToRgb(hue, 1, 1);
	$: satGradient = `linear-gradient(to top, #808080 0%, rgb(${chroma.r},${chroma.g},${chroma.b}) 100%)`;
	$: satNorm = satPct / 100;
	$: valMid = hsvToRgb(hue, satNorm, 1);
	$: valGradient =
		satPct <= 0
			? `linear-gradient(to top, #000000 0%, #ffffff 100%)`
			: `linear-gradient(to top, #000000 0%, rgb(${valMid.r},${valMid.g},${valMid.b}) 50%, #ffffff 100%)`;

	const trackLen = 200;
</script>

<div
	class="flag-color-vpicker flex flex-row items-stretch justify-center gap-3 sm:gap-5 touch-manipulation select-none"
	aria-label="HSB color controls"
>
	<!-- Hue -->
	<div class="flex flex-col items-center gap-2 w-14 shrink-0">
		<div
			class="relative flex items-center justify-center rounded-full border border-white/15 shadow-inner bg-neutral-900 overflow-hidden"
			style="width: 44px; height: {trackLen}px;"
		>
			<div
				class="pointer-events-none absolute inset-1 rounded-full"
				style="background: linear-gradient(to top, #f00 0%, #ff0 17%, #0f0 33%, #0ff 50%, #00f 67%, #f0f 83%, #f00 100%);"
			></div>
			<input
				type="range"
				min="0"
				max="359"
				bind:value={hue}
				{disabled}
				class="absolute z-10 cursor-pointer opacity-0"
				style="width: {trackLen}px; height: 44px; transform: rotate(-90deg);"
				aria-label="Hue"
			/>
			<div
				class="pointer-events-none absolute left-1/2 -translate-x-1/2 w-[85%] h-1 bg-white rounded-full shadow-md ring-1 ring-black/50 z-[5]"
				style="bottom: calc(8px + (100% - 16px) * {hue / 359});"
			></div>
		</div>
		<span class="text-[10px] uppercase tracking-wider text-sandy-light font-semibold">Hue</span>
	</div>

	<!-- Saturation -->
	<div class="flex flex-col items-center gap-2 w-14 shrink-0">
		<div
			class="relative flex items-center justify-center rounded-full border border-white/15 shadow-inner overflow-hidden"
			style="width: 44px; height: {trackLen}px;"
		>
			<div class="pointer-events-none absolute inset-1 rounded-full" style="background: {satGradient};"></div>
			<input
				type="range"
				min="0"
				max="100"
				bind:value={satPct}
				{disabled}
				class="absolute z-10 cursor-pointer opacity-0"
				style="width: {trackLen}px; height: 44px; transform: rotate(-90deg);"
				aria-label="Saturation"
			/>
			<div
				class="pointer-events-none absolute left-1/2 -translate-x-1/2 w-[85%] h-1 bg-white rounded-full shadow-md ring-1 ring-black/50 z-[5]"
				style="bottom: calc(8px + (100% - 16px) * {satPct / 100});"
			></div>
		</div>
		<span class="text-[10px] uppercase tracking-wider text-sandy-light font-semibold">Sat.</span>
	</div>

	<!-- Brightness -->
	<div class="flex flex-col items-center gap-2 w-14 shrink-0">
		<div
			class="relative flex items-center justify-center rounded-full border border-white/15 shadow-inner overflow-hidden"
			style="width: 44px; height: {trackLen}px;"
		>
			<div class="pointer-events-none absolute inset-1 rounded-full" style="background: {valGradient};"></div>
			<input
				type="range"
				min="0"
				max="100"
				bind:value={valPct}
				{disabled}
				class="absolute z-10 cursor-pointer opacity-0"
				style="width: {trackLen}px; height: 44px; transform: rotate(-90deg);"
				aria-label="Brightness"
			/>
			<div
				class="pointer-events-none absolute left-1/2 -translate-x-1/2 w-[85%] h-1 bg-white rounded-full shadow-md ring-1 ring-black/50 z-[5]"
				style="bottom: calc(8px + (100% - 16px) * {valPct / 100});"
			></div>
		</div>
		<span class="text-[10px] uppercase tracking-wider text-sandy-light font-semibold">Bright.</span>
	</div>
</div>

<style>
	:global(.flag-color-vpicker input[type='range']) {
		-webkit-appearance: none;
		appearance: none;
		background: transparent;
	}
	:global(.flag-color-vpicker input[type='range']::-webkit-slider-thumb) {
		-webkit-appearance: none;
		appearance: none;
		width: 28px;
		height: 28px;
		opacity: 0;
	}
	:global(.flag-color-vpicker input[type='range']::-moz-range-thumb) {
		width: 28px;
		height: 28px;
		opacity: 0;
		border: none;
	}
</style>
