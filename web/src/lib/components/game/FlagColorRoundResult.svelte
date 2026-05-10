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
</script>

<div class="relative w-full max-w-lg mx-auto rounded-[28px] overflow-hidden border border-white/15 shadow-2xl">
	<p
		class="absolute top-3 left-1/2 -translate-x-1/2 z-20 text-[11px] font-bold uppercase tracking-widest px-3 py-1 rounded-full bg-black/35 text-white/90 backdrop-blur-sm border border-white/10"
	>
		{roundCurrent}/{roundTotal}
	</p>

	<!-- Your selection -->
	<section
		class="relative pt-11 pb-8 px-5 sm:px-7 min-h-[140px] flex flex-col gap-3 sm:flex-row sm:justify-between sm:items-start"
		style="background-color: {guessHex};"
	>
		<div class="result-readout-panel w-full sm:w-auto sm:max-w-[min(100%,260px)]">
			<p class="text-[11px] font-bold uppercase tracking-[0.2em] mb-2 text-white/85">
				{labelYourSelection}
			</p>
			<p class="text-sm sm:text-base font-mono font-semibold text-white mb-2">{guessHsb}</p>
			<p class="text-[11px] font-mono text-white/75">{guessHex}</p>
		</div>
		<div class="result-readout-panel w-full sm:w-auto text-left sm:text-right sm:min-w-[9rem]">
			<p class="text-4xl sm:text-5xl font-black tabular-nums leading-none tracking-tight">{scoreTen}</p>
			<p class="text-xs sm:text-sm font-semibold mt-2 max-w-[12rem] sm:max-w-[10rem] sm:ml-auto leading-snug text-white/90">
				{tierMessage}
			</p>
			<p class="text-[10px] sm:text-xs font-mono mt-2 text-white/70">ΔE {deltaE.toFixed(1)}</p>
		</div>
	</section>

	<!-- Original -->
	<section class="relative pt-8 pb-14 px-5 sm:px-7 min-h-[130px]" style="background-color: {correctHex};">
		<div class="result-readout-panel inline-block max-w-full">
			<p class="text-[11px] font-bold uppercase tracking-[0.2em] mb-2 text-white/85">{labelOriginal}</p>
			<p class="text-sm sm:text-base font-mono font-semibold text-white mb-2">{correctHsb}</p>
			<p class="text-[11px] font-mono text-white/75">{correctHex}</p>
		</div>
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

<style lang="postcss">
	.result-readout-panel {
		@apply rounded-2xl bg-black/55 backdrop-blur-md border border-white/20 px-4 py-3 text-white shadow-lg;
	}
</style>
