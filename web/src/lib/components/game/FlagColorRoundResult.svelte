<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export let roundCurrent = 1;
	export let roundTotal = 10;
	export let scoreTen = '0.00';
	export let deltaE = 0;
	export let tierMessage = '';
	export let guessHsb = '';
	export let correctHsb = '';
	export let guessHex = '';
	export let correctHex = '';
	export let labelYourSelection = 'Your selection';
	export let labelOriginal = 'Original';

	const dispatch = createEventDispatcher<{ continue: void }>();

	function contrastFg(hex: string): string {
		const h = hex.replace('#', '');
		if (h.length !== 6) return 'rgba(255,255,255,0.95)';
		const r = parseInt(h.slice(0, 2), 16);
		const g = parseInt(h.slice(2, 4), 16);
		const b = parseInt(h.slice(4, 6), 16);
		const L = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
		return L > 0.55 ? 'rgba(15,23,42,0.92)' : 'rgba(255,255,255,0.95)';
	}

	$: fgGuess = contrastFg(guessHex);
	$: fgCorrect = contrastFg(correctHex);
</script>

<div class="relative w-full max-w-lg mx-auto rounded-[28px] overflow-hidden border border-white/15 shadow-2xl">
	<p
		class="absolute top-3 left-1/2 -translate-x-1/2 z-20 text-[11px] font-bold uppercase tracking-widest px-3 py-1 rounded-full bg-black/35 text-white/90 backdrop-blur-sm border border-white/10"
	>
		{roundCurrent}/{roundTotal}
	</p>

	<!-- Your selection -->
	<section
		class="relative pt-11 pb-8 px-5 sm:px-7 min-h-[140px]"
		style="background-color: {guessHex}; color: {fgGuess};"
	>
		<p class="text-[11px] font-bold uppercase tracking-[0.2em] opacity-80 mb-2">
			{labelYourSelection}
		</p>
		<p class="text-sm sm:text-base font-mono font-semibold opacity-95 mb-4">{guessHsb}</p>
		<p class="text-[11px] font-mono opacity-70">{guessHex}</p>
		<div class="absolute top-10 right-5 sm:right-7 text-right">
			<p class="text-4xl sm:text-5xl font-black tabular-nums leading-none tracking-tight">{scoreTen}</p>
			<p class="text-xs sm:text-sm font-semibold opacity-85 mt-2 max-w-[10rem] ml-auto leading-snug">
				{tierMessage}
			</p>
			<p class="text-[10px] sm:text-xs font-mono opacity-65 mt-2">ΔE {deltaE.toFixed(1)}</p>
		</div>
	</section>

	<!-- Original -->
	<section
		class="relative pt-8 pb-14 px-5 sm:px-7 min-h-[130px]"
		style="background-color: {correctHex}; color: {fgCorrect};"
	>
		<p class="text-[11px] font-bold uppercase tracking-[0.2em] opacity-80 mb-2">{labelOriginal}</p>
		<p class="text-sm sm:text-base font-mono font-semibold opacity-95 mb-2">{correctHsb}</p>
		<p class="text-[11px] font-mono opacity-75">{correctHex}</p>
	</section>

	<button
		type="button"
		class="absolute bottom-4 right-4 z-30 flex h-12 w-12 items-center justify-center rounded-full bg-black/55 hover:bg-black/70 border border-white/25 text-white text-xl shadow-lg backdrop-blur-sm transition-colors"
		on:click={() => dispatch('continue')}
		aria-label="Next round"
	>
		→
	</button>
</div>
