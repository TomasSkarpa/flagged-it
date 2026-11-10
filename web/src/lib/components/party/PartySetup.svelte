<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import DifficultySelector from '$lib/components/game/DifficultySelector.svelte';
	import RegionDropdown from '$lib/components/ui/RegionDropdown.svelte';
	import type { PartyGame, PartyMode } from '$lib/api/party';

	export let name = '';
	export let isLoading = false;
	export let error: string | null = null;

	let game: PartyGame = 'mix';
	let mode: PartyMode = 'live';
	let region = '';
	let joinCode = '';

	const dispatch = createEventDispatcher<{
		create: { mode: PartyMode; game: PartyGame; region: string; name: string };
		join: { code: string; name: string };
	}>();

	$: L = $locale;
	$: games = [
		{ id: 'mix', icon: '🎲', label: t('party.game.mix', undefined, L) },
		{ id: 'flag', icon: '🚩', label: t('party.game.flag', undefined, L) },
		{ id: 'shape', icon: '🗺️', label: t('party.game.shape', undefined, L) },
		{ id: 'capital', icon: '🏛️', label: t('party.game.capital', undefined, L) },
		{ id: 'compare', icon: '↕️', label: t('party.game.compare', undefined, L) },
		{ id: 'color', icon: '🎨', label: t('party.game.color', undefined, L) },
		{ id: 'fact', icon: '📚', label: t('party.game.fact', undefined, L) }
	] as { id: PartyGame; icon: string; label: string }[];
	$: modes = [
		{ value: 'live', icon: '⚡', label: t('party.mode.live', undefined, L), description: t('party.mode.live.desc', undefined, L) },
		{ value: 'challenge', icon: '🔗', label: t('party.mode.challenge', undefined, L), description: t('party.mode.challenge.desc', undefined, L) }
	];
	$: regions = [
		{ value: '', label: t('region.world', undefined, L) },
		{ value: 'Africa', label: t('region.africa', undefined, L) },
		{ value: 'Americas', label: t('region.americas', undefined, L) },
		{ value: 'Asia', label: t('region.asia', undefined, L) },
		{ value: 'Europe', label: t('region.europe', undefined, L) },
		{ value: 'Oceania', label: t('region.oceania', undefined, L) }
	];

	function create() {
		dispatch('create', { mode, game, region, name: name.trim() });
	}

	function join() {
		const code = joinCode.trim().toUpperCase();
		if (code.length >= 4) dispatch('join', { code, name: name.trim() });
	}
</script>

<div class="text-center px-4 md:px-8">
	<h1 class="text-4xl md:text-6xl font-bold text-sandy-light mb-4">
		<span class="mr-2 emoji-blue">👥</span>
		<span class="gradient-text">{t('party.title', undefined, L)}</span>
	</h1>
	<p class="text-lg text-text-muted mb-8">{t('party.description', undefined, L)}</p>

	<div class="card-game max-w-2xl mx-auto mb-6 text-left">
		<label class="block text-sm font-semibold text-sandy mb-2" for="party-name">{t('party.name.label', undefined, L)}</label>
		<input
			id="party-name"
			type="text"
			bind:value={name}
			maxlength="16"
			placeholder={t('party.name.placeholder', undefined, L)}
			autocomplete="nickname"
			class="party-input w-full mb-6"
		/>

		<p class="block text-sm font-semibold text-sandy mb-2">{t('party.game.label', undefined, L)}</p>
		<div class="flex flex-wrap gap-2 mb-6">
			{#each games as g}
				<button
					type="button"
					class="px-4 py-2 rounded-full border-2 font-semibold transition-all duration-200
						{game === g.id ? 'bg-primary/30 border-primary text-sandy-light' : 'border-white/20 text-sandy-light hover:border-accent'}"
					aria-pressed={game === g.id}
					on:click={() => (game = g.id)}
				>
					<span aria-hidden="true">{g.icon}</span>
					{g.label}
				</button>
			{/each}
		</div>

		<RegionDropdown {regions} bind:selected={region} label={t('game.setup.select_region', undefined, L)} />
	</div>

	<div class="max-w-2xl mx-auto mb-8">
		<DifficultySelector options={modes} selected={mode} on:select={(e) => (mode = e.detail.value === 'challenge' ? 'challenge' : 'live')} />
	</div>

	<button class="btn-primary px-12 py-4 text-xl font-bold" disabled={isLoading} on:click={create}>
		{isLoading ? t('party.creating', undefined, L) : t('party.create', undefined, L)}
	</button>

	{#if error}
		<div class="mt-6 p-4 bg-error/20 border border-error rounded-lg max-w-md mx-auto">
			<p class="text-error font-semibold">{error}</p>
		</div>
	{/if}

	<form class="max-w-md mx-auto mt-12" on:submit|preventDefault={join}>
		<p class="text-text-muted font-semibold mb-3">{t('party.join.title', undefined, L)}</p>
		<div class="flex gap-3">
			<input
				type="text"
				bind:value={joinCode}
				maxlength="4"
				placeholder={t('party.join.placeholder', undefined, L)}
				autocomplete="off"
				autocapitalize="characters"
				class="party-input flex-1 min-w-0 text-center uppercase tracking-[0.3em] font-bold"
				aria-label={t('party.join.placeholder', undefined, L)}
			/>
			<button type="submit" class="btn-secondary px-6 py-3 font-semibold" disabled={isLoading || joinCode.trim().length < 4}>
				{t('party.join.button', undefined, L)}
			</button>
		</div>
	</form>
</div>

<style>
	:global(.party-input) {
		padding: 0.75rem 1rem;
		border-radius: 0.5rem;
		border: 2px solid rgb(255 255 255 / 0.2);
		background: rgb(255 255 255 / 0.05);
		color: var(--color-sandy-light);
		transition: border-color 0.2s;
	}
	:global(.party-input:focus) {
		outline: none;
		border-color: var(--color-primary);
	}
	:global(:root.light .party-input) {
		border-color: rgb(15 23 42 / 0.2);
		background: #fff;
		color: #0f172a;
	}
</style>
