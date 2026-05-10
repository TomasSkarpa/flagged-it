<script lang="ts">
	export let hue = 200;
	export let satPct = 85;
	export let valPct = 75;

	export let disabled = false;

	function clampHue(n: number): number {
		if (!Number.isFinite(n)) return hue;
		let h = Math.round(n) % 360;
		if (h < 0) h += 360;
		return h;
	}

	function clampPct(n: number): number {
		if (!Number.isFinite(n)) return 0;
		return Math.max(0, Math.min(100, Math.round(n)));
	}

	function onHueInput(ev: Event): void {
		hue = clampHue(+(ev.currentTarget as HTMLInputElement).value);
	}

	function onSatInput(ev: Event): void {
		satPct = clampPct(+(ev.currentTarget as HTMLInputElement).value);
	}

	function onValInput(ev: Event): void {
		valPct = clampPct(+(ev.currentTarget as HTMLInputElement).value);
	}
</script>

<div class="flex flex-col gap-3 max-w-xs mx-auto touch-manipulation" aria-label="HSB numeric inputs">
	<label class="flex flex-col gap-1">
		<span class="text-[10px] uppercase tracking-wider text-sandy-light font-semibold">Hue (0–359)</span>
		<input
			type="number"
			min="0"
			max="359"
			step="1"
			value={Math.round(((hue % 360) + 360) % 360)}
			{disabled}
			class="px-3 py-2 rounded-xl bg-neutral-900 border border-white/15 text-white font-mono text-sm"
			on:input={onHueInput}
		/>
	</label>
	<label class="flex flex-col gap-1">
		<span class="text-[10px] uppercase tracking-wider text-sandy-light font-semibold">Saturation (0–100)</span>
		<input
			type="number"
			min="0"
			max="100"
			step="1"
			value={satPct}
			{disabled}
			class="px-3 py-2 rounded-xl bg-neutral-900 border border-white/15 text-white font-mono text-sm"
			on:input={onSatInput}
		/>
	</label>
	<label class="flex flex-col gap-1">
		<span class="text-[10px] uppercase tracking-wider text-sandy-light font-semibold">Brightness (0–100)</span>
		<input
			type="number"
			min="0"
			max="100"
			step="1"
			value={valPct}
			{disabled}
			class="px-3 py-2 rounded-xl bg-neutral-900 border border-white/15 text-white font-mono text-sm"
			on:input={onValInput}
		/>
	</label>
</div>
