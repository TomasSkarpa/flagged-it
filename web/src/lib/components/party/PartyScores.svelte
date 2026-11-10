<script lang="ts">
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import type { PartyPlayer } from '$lib/api/party';

	export let players: PartyPlayer[] = [];
	/** What to show on the right of each row. */
	export let show: 'none' | 'score' | 'answered' | 'round' | 'challenge' = 'score';
	export let ranked = true;

	$: L = $locale;
	const medals = ['🥇', '🥈', '🥉'];
</script>

<ul class="flex flex-col gap-2">
	{#each players as p, i (p.id)}
		<li
			class="flex items-center gap-3 px-4 py-2.5 rounded-lg border-2 transition-all duration-200
				{p.isYou ? 'border-primary bg-primary/10' : 'border-white/10 bg-white/5'}
				{p.connected || show === 'challenge' ? '' : 'opacity-50'}"
		>
			{#if ranked}
				<span class="w-7 text-center font-bold text-sandy stat-number">{medals[i] ?? i + 1}</span>
			{/if}
			<span class="flex-1 min-w-0 truncate font-semibold text-sandy-light">
				{p.name}
				{#if p.isYou}<span class="party-tag">{t('party.you', undefined, L)}</span>{/if}
				{#if p.isHost && show !== 'challenge'}<span class="party-tag">{t('party.host', undefined, L)}</span>{/if}
			</span>

			{#if show === 'answered'}
				<span class="text-lg" aria-label={p.answered ? '✓' : '…'}>{p.answered ? '✅' : '⏳'}</span>
			{:else if show === 'round'}
				{#if p.points !== undefined}
					<span class="text-sm font-bold stat-number {p.points > 0 ? 'text-success' : 'text-text-muted'}">+{p.points}</span>
				{/if}
			{:else if show === 'challenge' && !p.finished}
				<span class="text-sm text-text-muted italic">{t('party.challenge.playing', undefined, L)}</span>
			{/if}

			{#if show !== 'none' && show !== 'answered'}
				<span class="font-bold stat-number text-sandy-light min-w-[4.5rem] text-right">
					{p.score} <span class="text-xs text-text-muted font-normal">{t('party.points', undefined, L)}</span>
				</span>
			{/if}
		</li>
	{/each}
</ul>

<style>
	.party-tag {
		margin-left: 0.4rem;
		padding: 0.05rem 0.45rem;
		border-radius: 999px;
		font-size: 0.7rem;
		font-weight: 600;
		vertical-align: middle;
		background: rgb(255 255 255 / 0.1);
		color: var(--color-sandy);
	}
	:global(:root.light) .party-tag {
		background: rgb(15 23 42 / 0.08);
	}
</style>
