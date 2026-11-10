<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import type { PartyView } from '$lib/api/party';
	import PartyScores from './PartyScores.svelte';
	import ShareLink from './ShareLink.svelte';

	export let view: PartyView;
	export let shareUrl: string;
	export let isLoading = false;

	const dispatch = createEventDispatcher<{ start: void; leave: void }>();

	$: L = $locale;
	$: host = view.players.find((p) => p.isHost);
</script>

<div class="max-w-xl mx-auto text-center">
	<div class="card-game mb-6">
		<p class="text-sm font-semibold text-sandy uppercase tracking-wide mb-2">{t('party.lobby.code', undefined, L)}</p>
		<p class="party-code gradient-text mb-3" data-testid="room-code">{view.code}</p>
		<p class="text-text-muted mb-4">{t('party.lobby.share', undefined, L)}</p>
		<ShareLink url={shareUrl} />
	</div>

	<div class="card-game mb-6 text-left">
		<h2 class="text-lg font-bold text-sandy-light mb-3 mt-0">
			{t('party.players', undefined, L)} · {view.players.length}
		</h2>
		<PartyScores players={view.players} show="none" ranked={false} />
	</div>

	{#if view.you.isHost}
		<button class="btn-primary px-12 py-4 text-xl font-bold" disabled={isLoading} on:click={() => dispatch('start')}>
			{t('party.lobby.start', undefined, L)}
		</button>
		{#if view.players.length < 2}
			<p class="text-text-muted mt-3">{t('party.lobby.alone', undefined, L)}</p>
		{/if}
	{:else}
		<p class="text-lg text-text-muted">
			{t('party.lobby.waiting_host', { Name: host?.name ?? '' }, L)}
		</p>
	{/if}

	<button class="block mx-auto mt-8 text-text-muted underline text-sm" on:click={() => dispatch('leave')}>
		{t('party.leave', undefined, L)}
	</button>
</div>

<style>
	.party-code {
		font-size: clamp(3rem, 12vw, 4.5rem);
		font-weight: 800;
		letter-spacing: 0.25em;
		line-height: 1.1;
		padding-left: 0.25em; /* balance the letter spacing */
	}
</style>
