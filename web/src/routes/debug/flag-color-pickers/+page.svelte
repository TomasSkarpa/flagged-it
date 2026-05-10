<script lang="ts">
	import { goto } from '$app/navigation';
	import FlagColorVerticalPicker from '$lib/components/game/FlagColorVerticalPicker.svelte';
	import FlagColorHorizontalPicker from '$lib/components/game/FlagColorHorizontalPicker.svelte';
	import FlagColorSvPlanePicker from '$lib/components/game/FlagColorSvPlanePicker.svelte';
	import FlagColorRgbSliders from '$lib/components/game/FlagColorRgbSliders.svelte';
	import FlagColorHsbNumeric from '$lib/components/game/FlagColorHsbNumeric.svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import { formatHsbLabel, hexToRgb, hsvToRgb, rgbToHex, rgbToHsv } from '$lib/utils/colorHsv';

	$: currentLocale = $locale;
	$: pageTitle = t('debug.flag_color_pickers.document_title', undefined, currentLocale);
	$: pageDescription = t('debug.flag_color_pickers.meta_description', undefined, currentLocale);

	let hue = 200;
	let satPct = 85;
	let valPct = 75;

	$: sat = satPct / 100;
	$: val = valPct / 100;
	$: rgbPreview = hsvToRgb(hue, sat, val);
	$: hexValue = rgbToHex(rgbPreview.r, rgbPreview.g, rgbPreview.b);
	$: hsbLabel = formatHsbLabel(hue, sat, val);

	function reset(): void {
		hue = 200;
		satPct = 85;
		valPct = 75;
	}

	function onNativeColorInput(ev: Event): void {
		const el = ev.currentTarget as HTMLInputElement;
		const rgb = hexToRgb(el.value);
		const hsv = rgbToHsv(rgb.r, rgb.g, rgb.b);
		hue = Math.round(((hsv.h % 360) + 360) % 360);
		satPct = Math.round(hsv.s * 100);
		valPct = Math.round(hsv.v * 100);
	}
</script>

<svelte:head>
	<title>{pageTitle}</title>
	<meta name="description" content={pageDescription} />
</svelte:head>

<div class="min-h-screen p-4 md:p-8 pb-16">
	<div class="max-w-5xl mx-auto space-y-8">
		<header class="space-y-3">
			<h1 class="text-3xl md:text-4xl font-bold text-sandy-light">
				<span class="emoji-blue mr-2">🎨</span>
				Flag color picker experiments
			</h1>
			<p class="text-text-muted max-w-2xl">
				Shared HSB state across variants below (production uses the vertical HSB control). Use this page to
				compare layouts and add new picker implementations.
			</p>
			<div class="flex flex-wrap gap-3">
				<button type="button" class="btn-primary px-4 py-2 rounded-xl font-semibold" on:click={() => goto('/flag-color-game')}>
					Open flag color game
				</button>
				<button type="button" class="btn-secondary px-4 py-2 rounded-xl font-semibold" on:click={() => goto('/debug')}>
					Debug hub
				</button>
				<button type="button" class="btn-secondary px-4 py-2 rounded-xl font-semibold" on:click={reset}>
					Reset HSB
				</button>
			</div>
		</header>

		<section class="card-game p-5 md:p-6 space-y-4">
			<h2 class="text-lg font-bold text-sandy-light">Current selection</h2>
			<div class="flex flex-col sm:flex-row gap-4 items-stretch sm:items-center">
				<div
					class="w-full sm:w-40 h-24 rounded-2xl border border-white/15 shadow-inner shrink-0"
					style="background-color: rgb({rgbPreview.r},{rgbPreview.g},{rgbPreview.b});"
					aria-hidden="true"
				></div>
				<dl class="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm font-mono text-white/90">
					<dt class="text-white/55">HSB</dt>
					<dd>{hsbLabel}</dd>
					<dt class="text-white/55">RGB</dt>
					<dd>{rgbPreview.r}, {rgbPreview.g}, {rgbPreview.b}</dd>
					<dt class="text-white/55">Hex</dt>
					<dd>{hexValue}</dd>
				</dl>
			</div>
		</section>

		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<section class="card-game p-5 md:p-6 flex flex-col gap-4">
				<div>
					<h2 class="text-xl font-bold text-sandy-light">Vertical HSB (production)</h2>
					<p class="text-sm text-text-muted mt-1">
						Same component as <code class="text-white/80">FlagColorVerticalPicker</code> on the live game.
					</p>
				</div>
				<div class="flex justify-center py-2">
					<FlagColorVerticalPicker bind:hue bind:satPct bind:valPct />
				</div>
			</section>

			<section class="card-game p-5 md:p-6 flex flex-col gap-4">
				<div>
					<h2 class="text-xl font-bold text-sandy-light">Horizontal HSB (debug)</h2>
					<p class="text-sm text-text-muted mt-1">
						Same model as vertical; gradients and thumbs run left-to-right for comparison.
					</p>
				</div>
				<div class="flex justify-center py-2">
					<FlagColorHorizontalPicker bind:hue bind:satPct bind:valPct />
				</div>
			</section>

			<section class="card-game p-5 md:p-6 flex flex-col gap-4">
				<div>
					<h2 class="text-xl font-bold text-sandy-light">Native color input</h2>
					<p class="text-sm text-text-muted mt-1">
						Browser <code class="text-white/80">type="color"</code>; updates shared HSB when changed.
					</p>
				</div>
				<div class="flex flex-col sm:flex-row items-center gap-6 py-4">
					<label class="flex flex-col items-center gap-2 cursor-pointer">
						<span class="text-xs uppercase tracking-wider text-white/55 font-semibold">Picker</span>
						<input
							type="color"
							class="w-16 h-16 rounded-xl border border-white/20 cursor-pointer bg-transparent"
							value={hexValue}
							on:input={onNativeColorInput}
							aria-label="Native color picker"
						/>
					</label>
					<p class="text-sm text-text-muted text-center sm:text-left">
						Hex round-trips through <code class="text-white/80">rgbToHsv</code>; slight drift is possible
						due to rounding.
					</p>
				</div>
			</section>

			<section class="card-game p-5 md:p-6 flex flex-col gap-4">
				<div>
					<h2 class="text-xl font-bold text-sandy-light">SV plane + hue</h2>
					<p class="text-sm text-text-muted mt-1">
						2D saturation (horizontal) and brightness (vertical) at the current hue; separate hue strip.
					</p>
				</div>
				<div class="flex justify-center py-2">
					<FlagColorSvPlanePicker bind:hue bind:satPct bind:valPct />
				</div>
			</section>

			<section class="card-game p-5 md:p-6 flex flex-col gap-4">
				<div>
					<h2 class="text-xl font-bold text-sandy-light">RGB sliders</h2>
					<p class="text-sm text-text-muted mt-1">
						sRGB 0–255 per channel; gradients reflect the other two channels. Writes back through <code
							class="text-white/80">rgbToHsv</code>.
					</p>
				</div>
				<div class="flex justify-center py-2">
					<FlagColorRgbSliders bind:hue bind:satPct bind:valPct />
				</div>
			</section>

			<section class="card-game p-5 md:p-6 flex flex-col gap-4">
				<div>
					<h2 class="text-xl font-bold text-sandy-light">HSB numeric</h2>
					<p class="text-sm text-text-muted mt-1">
						Direct entry for H°, saturation %, and brightness %. Useful for matching reported values.
					</p>
				</div>
				<div class="flex justify-center py-2">
					<FlagColorHsbNumeric bind:hue bind:satPct bind:valPct />
				</div>
			</section>
		</div>
	</div>
</div>
