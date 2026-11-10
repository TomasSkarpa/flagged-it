<script lang="ts">
	import { createEventDispatcher, onDestroy, onMount } from 'svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import type { PartyView } from '$lib/api/party';
	import { partyFlagUrl, fetchPartyShape } from '$lib/api/party';
	import { formatValue } from '$lib/api/higherLowerGame';
	import { hsvToRgb } from '$lib/utils/colorHsv';
	import ShapeRenderer from '$lib/components/ShapeRenderer.svelte';
	import FlagColorFlag from '$lib/components/game/FlagColorFlag.svelte';
	import FlagColorSvPlanePicker from '$lib/components/game/FlagColorSvPlanePicker.svelte';
	import PartyScores from './PartyScores.svelte';

	export let view: PartyView;
	export let code: string;
	export let token: string;
	export let isSubmitting = false;

	const dispatch = createEventDispatcher<{
		answer: { optionId?: string; r?: number; g?: number; b?: number };
		next: void;
	}>();

	$: L = $locale;
	$: q = view.question!;
	$: reveal = view.phase === 'reveal' ? view.reveal : undefined;
	$: mine = view.you.answer;
	$: live = view.mode === 'live';

	// Timer: the server sends time left; count down locally.
	let deadline = 0;
	let left = 0;
	$: duration = Math.max(view.durationMs, 1);
	$: {
		deadline = Date.now() + view.remainingMs;
		left = view.remainingMs;
	}
	let tick: ReturnType<typeof setInterval>;
	onMount(() => {
		tick = setInterval(() => (left = Math.max(0, deadline - Date.now())), 100);
	});
	onDestroy(() => clearInterval(tick));
	$: timeUp = view.phase === 'question' && left <= 0;
	$: seconds = Math.ceil(left / 1000);

	// Per-round local state, reset when the question changes.
	let roundKey = '';
	let picked: string | null = null;
	let imageLoaded = false;
	let shape: any = null;
	let hue = 200;
	let satPct = 85;
	let valPct = 75;
	$: if (q && q.key !== roundKey) {
		roundKey = q.key;
		picked = null;
		imageLoaded = false;
		shape = null;
		hue = Math.floor(Math.random() * 360);
		satPct = 70;
		valPct = 70;
		if (q.kind === 'shape') loadShape(q.key, view.round);
	}
	$: rgb = hsvToRgb(hue, satPct / 100, valPct / 100);

	async function loadShape(key: string, round: number) {
		try {
			const geo = await fetchPartyShape(code, token, round);
			if (roundKey === key) shape = geo;
		} catch {
			/* the round still works without the picture loading */
		}
	}

	$: locked = !!mine || picked !== null || view.phase !== 'question' || timeUp;
	$: answeredCount = view.players.filter((p) => p.answered).length;

	function choose(id: string) {
		if (locked || isSubmitting) return;
		picked = id;
		dispatch('answer', { optionId: id });
	}

	function lockColor() {
		if (locked || isSubmitting) return;
		picked = 'color';
		dispatch('answer', { r: rgb.r, g: rgb.g, b: rgb.b });
	}

	$: yourPick = mine?.optionId ?? (picked !== 'color' ? picked : null);

	function optionClass(id: string): string {
		if (reveal) {
			if (id === reveal.answerId) return 'bg-success border-success text-white shadow-glow';
			if (id === yourPick) return 'bg-error border-error text-white';
			return 'border-white/10 text-sandy-light opacity-60';
		}
		if (id === yourPick) return 'bg-primary/30 border-primary text-sandy-light';
		return 'answer-grid-btn border-white/20 text-sandy-light hover:border-accent hover:bg-[var(--color-surface-light)]';
	}

	function factHtml(text: string): string {
		const esc = text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
		return esc.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
	}

	function compareValue(id: string): string {
		const v = reveal?.values?.[id];
		if (v === undefined) return '';
		const metric = q.metric === 'area' ? 'area' : 'population';
		return formatValue(v, metric) + (metric === 'area' ? ' km²' : '');
	}

	$: questionText =
		q.kind === 'flag'
			? t('party.q.flag', undefined, L)
			: q.kind === 'shape'
				? t('party.q.shape', undefined, L)
				: q.kind === 'capital'
					? t('party.q.capital', { Country: q.countryName ?? '' }, L)
					: q.kind === 'compare'
						? t(q.metric === 'area' ? 'party.q.compare.area' : 'party.q.compare.population', undefined, L)
						: q.kind === 'color'
							? t('party.q.color', { Country: q.countryName ?? '' }, L)
							: t('party.q.fact', undefined, L);

	$: resultText = !reveal
		? ''
		: q.kind === 'color'
			? mine && !mine.timedOut
				? t('party.reveal.color', { Points: mine.points }, L)
				: t('party.reveal.color_timeout', undefined, L)
			: mine?.correct
				? t('party.reveal.correct', { Points: mine.points }, L)
				: !mine || mine.timedOut
					? t('party.reveal.timeout', { Answer: reveal.answerLabel ?? '' }, L)
					: t('party.reveal.wrong', { Answer: reveal.answerLabel ?? '' }, L);
	$: resultGood = !!mine && (q.kind === 'color' ? mine.points > 0 : mine.correct);
</script>

<div class="mb-2">
	<div class="flex items-center justify-between gap-2 mb-2">
		<div class="text-base text-text-light font-semibold stat-number round-counter">
			{t('game.round.progress', [view.round, view.rounds], L)}
		</div>
		<div class="flex items-center gap-3">
			{#if view.phase === 'question'}
				<span class="font-bold stat-number {seconds <= 5 ? 'text-error' : 'text-sandy-light'}" data-testid="timer">⏱ {seconds}s</span>
			{/if}
			<span class="font-bold stat-number text-sandy-light">{view.you.score} {t('party.points', undefined, L)}</span>
		</div>
	</div>
	<div class="w-full h-2 bg-white/10 rounded-full overflow-hidden border border-white/20">
		<div
			class="h-full rounded-full {left / duration < 0.33 ? 'bg-error' : 'bg-terracotta'}"
			style="width: {view.phase === 'question' ? (left / duration) * 100 : 0}%; transition: width 0.1s linear"
		></div>
	</div>
</div>

<div class="card-game relative overflow-hidden mt-4" data-testid="round-card" data-kind={q.kind}>
	<h2 class="text-xl md:text-2xl font-bold text-sandy-light text-center mb-4 mt-2">{questionText}</h2>

	{#if q.kind === 'flag'}
		<div class="relative w-full max-w-72 mx-auto mb-4 flex items-center justify-center">
			{#if !imageLoaded}
				<div class="absolute inset-0 bg-white/5 rounded-[25px] animate-pulse border border-white/10"></div>
			{/if}
			<img
				src={partyFlagUrl(code, token, view.round)}
				alt=""
				class="w-full h-auto relative transition-opacity duration-200 {imageLoaded ? 'opacity-100' : 'opacity-0'}"
				on:load={() => (imageLoaded = true)}
				on:error={() => (imageLoaded = true)}
			/>
		</div>
	{:else if q.kind === 'shape'}
		<div class="party-shape mb-4">
			{#if shape}
				<ShapeRenderer geoJson={shape} width={320} height={180} fillColor="var(--color-sandy-light)" />
			{:else}
				<div class="w-full max-w-[320px] h-[180px] mx-auto bg-white/5 rounded-lg animate-pulse"></div>
			{/if}
		</div>
	{:else if q.kind === 'capital' && q.flagUrl}
		<img src={q.flagUrl} alt="" class="w-24 h-auto mx-auto mb-4" />
	{:else if q.kind === 'fact' && q.fact}
		<p class="party-fact text-lg text-sandy-light text-center max-w-2xl mx-auto mb-6 px-2">{@html factHtml(q.fact)}</p>
	{/if}

	{#if q.kind === 'compare'}
		<div class="grid grid-cols-2 gap-3 md:gap-4">
			{#each q.options ?? [] as o (o.id)}
				<button
					class="flex flex-col items-center gap-2 px-3 py-5 rounded-xl border-2 font-semibold transition-all duration-200 {optionClass(o.id)}"
					disabled={locked && !reveal}
					on:click={() => choose(o.id)}
				>
					{#if o.flagUrl}<img src={o.flagUrl} alt="" class="w-16 md:w-20 h-auto" />{/if}
					<span class="text-base md:text-lg text-center">{o.label}</span>
					{#if reveal}<span class="text-xl md:text-2xl font-bold stat-number">{compareValue(o.id)}</span>{/if}
				</button>
			{/each}
		</div>
	{:else if q.kind === 'color'}
		{#if !reveal}
			<div class="flex flex-col lg:flex-row items-center lg:items-start justify-center gap-6 pb-2">
				<FlagColorSvPlanePicker bind:hue bind:satPct bind:valPct disabled={locked} hueAriaLabel={t('game.flag_color.hue', undefined, L)} size="game" />
				<div class="flex flex-col items-stretch gap-4 w-full max-w-[280px]">
					<FlagColorFlag
						flagUrl={q.flagUrl ?? ''}
						guessableId={q.guessableId ?? ''}
						questionId={q.key}
						revealFullColor={false}
						previewR={rgb.r}
						previewG={rgb.g}
						previewB={rgb.b}
					/>
					<button class="btn-primary w-full px-8 py-3.5 font-semibold rounded-2xl disabled:opacity-50" disabled={locked || isSubmitting} on:click={lockColor}>
						{t('party.color.submit', undefined, L)}
					</button>
				</div>
			</div>
		{:else}
			<div class="flex flex-col sm:flex-row items-center justify-center gap-6">
				<div class="w-full max-w-[280px]">
					<FlagColorFlag flagUrl={q.flagUrl ?? ''} guessableId={q.guessableId ?? ''} questionId={q.key} revealFullColor={true} previewR={rgb.r} previewG={rgb.g} previewB={rgb.b} />
				</div>
				<div class="flex gap-4">
					{#if mine?.hex}
						<div class="text-center">
							<div class="w-16 h-16 rounded-xl border-2 border-white/30 mx-auto mb-1" style="background:{mine.hex}"></div>
							<span class="text-xs text-text-muted">{t('game.flag_color.result.your_selection', undefined, L)}</span>
						</div>
					{/if}
					<div class="text-center">
						<div class="w-16 h-16 rounded-xl border-2 border-white/30 mx-auto mb-1" style="background:{reveal.targetHex}"></div>
						<span class="text-xs text-text-muted">{t('game.flag_color.result.original', undefined, L)}</span>
					</div>
				</div>
			</div>
		{/if}
	{:else}
		<div class="grid gap-3 md:gap-4 grid-cols-1 md:grid-cols-2">
			{#each q.options ?? [] as o (o.id)}
				<button
					class="w-full px-6 py-4 rounded-lg border-2 font-semibold text-left transition-all duration-200 {optionClass(o.id)} disabled:cursor-not-allowed"
					disabled={locked}
					on:click={() => choose(o.id)}
				>
					{o.label}
				</button>
			{/each}
		</div>
	{/if}

	<div class="min-h-[3.5rem] mt-5 text-center" aria-live="polite">
		{#if reveal}
			<p class="text-xl font-bold {resultGood ? 'text-success' : 'text-error'}" data-testid="result">{resultText}</p>
			{#if !live}
				<button class="btn-primary px-10 py-3 mt-4 font-bold" disabled={isSubmitting} on:click={() => dispatch('next')}>
					{t('party.next', undefined, L)}
				</button>
			{/if}
		{:else if live && (mine || picked)}
			<p class="text-text-muted font-semibold">{t('party.answered', undefined, L)}</p>
			<p class="text-sm text-text-muted">{t('party.answered_count', { Done: answeredCount, Total: view.players.length }, L)}</p>
		{/if}
	</div>
</div>

{#if live}
	<div class="mt-6 max-w-xl mx-auto">
		<PartyScores players={view.players} show={reveal ? 'round' : 'answered'} />
	</div>
{/if}

<style>
	.party-shape {
		display: flex;
		justify-content: center;
	}
	.party-shape :global(svg) {
		width: 100%;
		max-width: 320px;
		max-height: 180px;
		height: auto;
	}
	.party-fact :global(strong) {
		color: var(--color-accent);
	}
	:global(:root.light) .round-counter {
		color: #0f172a !important;
	}
</style>
