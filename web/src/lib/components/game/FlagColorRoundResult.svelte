<script lang="ts">
	import { createEventDispatcher } from 'svelte';

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

<div
	class="flag-color-round-result relative w-full max-w-lg mx-auto rounded-[28px] overflow-hidden border border-white/15 shadow-2xl"
>
	<!-- Your selection -->
	<section
		class="relative pt-8 pb-8 px-5 sm:px-7 min-h-[140px] flex flex-col gap-3 sm:flex-row sm:justify-between sm:items-start"
		style="background-color: {guessHex};"
	>
		<div class="result-readout-panel w-full sm:w-auto sm:max-w-[min(100%,260px)]">
			<p class="result-readout-label">{labelYourSelection}</p>
			<p class="result-readout-mono">{guessHsb}</p>
			<p class="result-readout-hex">{guessHex}</p>
		</div>
		<div class="result-readout-panel w-full sm:w-auto text-left sm:text-right sm:min-w-[9rem]">
			<p class="result-readout-score">{scoreTen}</p>
			<p class="result-readout-tier">{tierMessage}</p>
			<p class="result-readout-delta">ΔE {deltaE.toFixed(1)}</p>
		</div>
	</section>

	<!-- Original -->
	<section class="relative pt-8 pb-14 px-5 sm:px-7 min-h-[130px]" style="background-color: {correctHex};">
		<div class="result-readout-panel inline-block max-w-full">
			<p class="result-readout-label">{labelOriginal}</p>
			<p class="result-readout-mono">{correctHsb}</p>
			<p class="result-readout-hex">{correctHex}</p>
		</div>
	</section>

	<button
		type="button"
		class="flag-color-round-result-next absolute bottom-4 right-4 z-30 flex h-12 w-12 items-center justify-center rounded-full bg-black/55 hover:bg-black/70 border border-white/25 text-white text-xl shadow-lg backdrop-blur-sm transition-colors"
		on:click={() => dispatch('continue')}
		aria-label="Next round"
	>
		→
	</button>
</div>

<style lang="postcss">
	.result-readout-panel {
		@apply rounded-2xl bg-black/55 backdrop-blur-md border border-white/20 px-4 py-3 shadow-lg;
	}

	.result-readout-label {
		@apply text-[11px] font-bold uppercase tracking-[0.2em] mb-2 text-white/85;
	}

	.result-readout-mono {
		@apply text-sm sm:text-base font-mono font-semibold text-white mb-2;
	}

	.result-readout-hex {
		@apply text-[11px] font-mono text-white/75;
	}

	.result-readout-score {
		@apply text-4xl sm:text-5xl font-black tabular-nums leading-none tracking-tight text-white;
	}

	.result-readout-tier {
		@apply text-xs sm:text-sm font-semibold mt-2 max-w-[12rem] sm:max-w-[10rem] sm:ml-auto leading-snug text-white/90;
	}

	.result-readout-delta {
		@apply text-[10px] sm:text-xs font-mono mt-2 text-white/70;
	}

	/* Light mode: frosted light panels + slate text (dark overlays read muddy on saturated bands). */
	:global(:root.light) .flag-color-round-result {
		border-color: rgba(15, 23, 42, 0.12);
		box-shadow:
			0 20px 40px rgba(15, 23, 42, 0.08),
			0 4px 12px rgba(15, 23, 42, 0.05);
	}

	:global(:root.light) .flag-color-round-result .result-readout-panel {
		background-color: rgba(255, 255, 255, 0.94);
		border-color: rgba(15, 23, 42, 0.12);
		box-shadow:
			0 4px 24px rgba(15, 23, 42, 0.07),
			0 1px 2px rgba(15, 23, 42, 0.05);
		backdrop-filter: blur(12px);
		-webkit-backdrop-filter: blur(12px);
	}

	:global(:root.light) .flag-color-round-result .result-readout-label {
		color: #64748b;
	}

	:global(:root.light) .flag-color-round-result .result-readout-mono,
	:global(:root.light) .flag-color-round-result .result-readout-score {
		color: #0f172a;
	}

	:global(:root.light) .flag-color-round-result .result-readout-tier {
		color: #334155;
	}

	:global(:root.light) .flag-color-round-result .result-readout-hex,
	:global(:root.light) .flag-color-round-result .result-readout-delta {
		color: #64748b;
	}

	:global(:root.light) .flag-color-round-result .flag-color-round-result-next {
		background-color: rgba(255, 255, 255, 0.96);
		border-color: rgba(15, 23, 42, 0.14);
		color: #0f172a;
		box-shadow:
			0 4px 16px rgba(15, 23, 42, 0.1),
			0 1px 3px rgba(15, 23, 42, 0.06);
		backdrop-filter: blur(8px);
		-webkit-backdrop-filter: blur(8px);
	}

	:global(:root.light) .flag-color-round-result .flag-color-round-result-next:hover {
		background-color: #ffffff;
	}
</style>
