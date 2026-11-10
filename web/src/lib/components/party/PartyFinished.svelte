<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import { triggerConfetti } from '$lib/utils/confetti';
	import type { PartyView } from '$lib/api/party';
	import PartyScores from './PartyScores.svelte';
	import ShareLink from './ShareLink.svelte';

	export let view: PartyView;
	export let shareUrl: string;
	export let isLoading = false;

	const dispatch = createEventDispatcher<{ again: void; leave: void; newGame: void }>();

	$: L = $locale;
	$: challenge = view.mode === 'challenge';
	$: finished = challenge ? view.players.filter((p) => p.finished) : view.players;
	$: winner = finished[0];
	$: youWon = !!winner && winner.isYou && (finished.length > 1 || !challenge);

	onMount(() => {
		if (winner?.isYou && view.players.length > 1) {
			setTimeout(() => triggerConfetti({ particleCount: 90, spread: 70, origin: { x: 0.5, y: 0.3 }, duration: 3000 }), 100);
		}
	});
</script>

<div class="max-w-xl mx-auto text-center">
	<div class="card-game mb-6">
		<div class="text-6xl mb-4">🏆</div>
		<h2 class="text-3xl md:text-4xl font-bold text-sandy-light mb-2">{t('party.finished.title', undefined, L)}</h2>
		{#if winner && view.players.length > 1}
			<p class="text-xl text-text-muted mb-6" data-testid="winner">
				{youWon ? t('party.finished.you_won', undefined, L) : t('party.finished.winner', { Name: winner.name }, L)}
			</p>
		{:else}
			<div class="mb-6"></div>
		{/if}
		<div class="text-left">
			<PartyScores players={view.players} show={challenge ? 'challenge' : 'score'} />
		</div>
	</div>

	{#if challenge}
		<div class="card-game mb-6 text-left">
			<p class="text-text-muted mb-3">{t('party.challenge.share', undefined, L)}</p>
			<ShareLink url={shareUrl} />
		</div>
		<button class="btn-primary px-10 py-4 text-lg font-bold" on:click={() => dispatch('newGame')}>
			{t('party.new_game', undefined, L)}
		</button>
	{:else if view.you.isHost}
		<button class="btn-primary px-12 py-4 text-xl font-bold" disabled={isLoading} on:click={() => dispatch('again')}>
			{t('party.again', undefined, L)}
		</button>
	{:else}
		<p class="text-lg text-text-muted">{t('party.waiting_again', undefined, L)}</p>
	{/if}

	{#if !challenge}
		<button class="block mx-auto mt-8 text-text-muted underline text-sm" on:click={() => dispatch('leave')}>
			{t('party.leave', undefined, L)}
		</button>
	{/if}
</div>
