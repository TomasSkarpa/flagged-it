<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import type { PartyView } from '$lib/api/party';
	import PartyScores from './PartyScores.svelte';

	export let view: PartyView;
	export let isLoading = false;

	const dispatch = createEventDispatcher<{ start: void }>();

	$: L = $locale;
	$: creator = view.players.find((p) => p.isHost);
	$: others = view.players.filter((p) => p.done > 0);
</script>

<div class="max-w-xl mx-auto text-center">
	<div class="card-game mb-6">
		<div class="text-6xl mb-4">🔗</div>
		<h2 class="text-3xl font-bold text-sandy-light mb-3">
			{t('party.challenge.from', { Name: creator && !creator.isYou ? creator.name : view.you.name }, L)}
		</h2>
		<p class="text-lg text-text-muted mb-6">{t('party.challenge.rules', { Rounds: view.rounds }, L)}</p>
		<button class="btn-primary px-12 py-4 text-xl font-bold" disabled={isLoading} on:click={() => dispatch('start')}>
			{t('party.challenge.start', undefined, L)}
		</button>
	</div>

	{#if others.length > 0}
		<div class="card-game text-left">
			<PartyScores players={others} show="challenge" />
		</div>
	{/if}
</div>
