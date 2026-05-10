<script lang="ts">
	import { hsvToRgb, rgbToHsv } from '$lib/utils/colorHsv';

	export let hue = 200;
	export let satPct = 85;
	export let valPct = 75;

	export let disabled = false;

	$: sat = satPct / 100;
	$: val = valPct / 100;
	$: rgb = hsvToRgb(hue, sat, val);

	function applyChannel(channel: 'r' | 'g' | 'b', next: number): void {
		const v = Math.max(0, Math.min(255, Math.round(next)));
		const r = channel === 'r' ? v : rgb.r;
		const g = channel === 'g' ? v : rgb.g;
		const b = channel === 'b' ? v : rgb.b;
		const hsv = rgbToHsv(r, g, b);
		hue = Math.round(((hsv.h % 360) + 360) % 360);
		satPct = Math.round(hsv.s * 100);
		valPct = Math.round(hsv.v * 100);
	}

	function onR(ev: Event): void {
		applyChannel('r', +(ev.currentTarget as HTMLInputElement).value);
	}
	function onG(ev: Event): void {
		applyChannel('g', +(ev.currentTarget as HTMLInputElement).value);
	}
	function onB(ev: Event): void {
		applyChannel('b', +(ev.currentTarget as HTMLInputElement).value);
	}

	const trackLen = 200;

	$: gradR = `linear-gradient(to right, rgb(0,${rgb.g},${rgb.b}) 0%, rgb(255,${rgb.g},${rgb.b}) 100%)`;
	$: gradG = `linear-gradient(to right, rgb(${rgb.r},0,${rgb.b}) 0%, rgb(${rgb.r},255,${rgb.b}) 100%)`;
	$: gradB = `linear-gradient(to right, rgb(${rgb.r},${rgb.g},0) 0%, rgb(${rgb.r},${rgb.g},255) 100%)`;
</script>

<div
	class="flag-color-rgbpick flex flex-col gap-4 sm:gap-5 touch-manipulation select-none max-w-md mx-auto"
	aria-label="RGB sliders"
>
	<div class="flex flex-col gap-2 w-full">
		<div
			class="relative flex items-center rounded-full border border-white/15 shadow-inner overflow-hidden mx-auto"
			style="width: {trackLen}px; height: 44px;"
		>
			<div class="pointer-events-none absolute inset-1 rounded-full" style="background: {gradR};"></div>
			<input
				type="range"
				min="0"
				max="255"
				value={rgb.r}
				{disabled}
				class="absolute z-10 cursor-pointer opacity-0 flag-color-rgb-range"
				style="width: {trackLen}px; height: 44px; left: 0; top: 0;"
				aria-label="RGB R"
				on:input={onR}
			/>
			<div
				class="pointer-events-none absolute top-1/2 -translate-y-1/2 h-[85%] w-1 bg-white rounded-full shadow-md ring-1 ring-black/50 z-[5]"
				style="left: calc(8px + (100% - 16px) * {rgb.r / 255});"
			></div>
		</div>
		<span class="text-[10px] uppercase tracking-wider text-white/75 font-semibold text-center">R</span>
	</div>

	<div class="flex flex-col gap-2 w-full">
		<div
			class="relative flex items-center rounded-full border border-white/15 shadow-inner overflow-hidden mx-auto"
			style="width: {trackLen}px; height: 44px;"
		>
			<div class="pointer-events-none absolute inset-1 rounded-full" style="background: {gradG};"></div>
			<input
				type="range"
				min="0"
				max="255"
				value={rgb.g}
				{disabled}
				class="absolute z-10 cursor-pointer opacity-0 flag-color-rgb-range"
				style="width: {trackLen}px; height: 44px; left: 0; top: 0;"
				aria-label="RGB G"
				on:input={onG}
			/>
			<div
				class="pointer-events-none absolute top-1/2 -translate-y-1/2 h-[85%] w-1 bg-white rounded-full shadow-md ring-1 ring-black/50 z-[5]"
				style="left: calc(8px + (100% - 16px) * {rgb.g / 255});"
			></div>
		</div>
		<span class="text-[10px] uppercase tracking-wider text-white/75 font-semibold text-center">G</span>
	</div>

	<div class="flex flex-col gap-2 w-full">
		<div
			class="relative flex items-center rounded-full border border-white/15 shadow-inner overflow-hidden mx-auto"
			style="width: {trackLen}px; height: 44px;"
		>
			<div class="pointer-events-none absolute inset-1 rounded-full" style="background: {gradB};"></div>
			<input
				type="range"
				min="0"
				max="255"
				value={rgb.b}
				{disabled}
				class="absolute z-10 cursor-pointer opacity-0 flag-color-rgb-range"
				style="width: {trackLen}px; height: 44px; left: 0; top: 0;"
				aria-label="RGB B"
				on:input={onB}
			/>
			<div
				class="pointer-events-none absolute top-1/2 -translate-y-1/2 h-[85%] w-1 bg-white rounded-full shadow-md ring-1 ring-black/50 z-[5]"
				style="left: calc(8px + (100% - 16px) * {rgb.b / 255});"
			></div>
		</div>
		<span class="text-[10px] uppercase tracking-wider text-white/75 font-semibold text-center">B</span>
	</div>
</div>

<style>
	:global(.flag-color-rgb-range) {
		-webkit-appearance: none;
		appearance: none;
		background: transparent;
	}
	:global(.flag-color-rgb-range::-webkit-slider-thumb) {
		-webkit-appearance: none;
		appearance: none;
		width: 28px;
		height: 28px;
		opacity: 0;
	}
	:global(.flag-color-rgb-range::-moz-range-thumb) {
		width: 28px;
		height: 28px;
		opacity: 0;
		border: none;
	}
</style>
