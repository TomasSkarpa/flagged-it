<script lang="ts">
	import { createEventDispatcher, onMount, onDestroy } from 'svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import {
		currentRoom,
		currentPlayers,
		currentQuestion,
		currentPlayerId,
		gameLeaderboard,
		gameFinished,
		finalScores,
		onMessage,
		submitAnswer
	} from '$lib/stores/multiplayer';
	import type { Player } from '$lib/stores/multiplayer';

	export let minimized: boolean = false;
	export let onExpand: (() => void) | null = null;

	const dispatch = createEventDispatcher<{
		answerSubmitted: { questionId: string; answer: string; timeTaken: number };
	}>();

	$: currentLocale = $locale;
	$: players = Object.values($currentPlayers);
	$: leaderboard = Object.values($gameLeaderboard).sort((a, b) => b.score - a.score);
	$: currentPlayer = $currentPlayers[$currentPlayerId || ''];
	$: playerRank = leaderboard.findIndex(p => p.id === $currentPlayerId) + 1;
	$: totalPlayers = leaderboard.length;

	let questionStartTime = Date.now();

	onMount(() => {
		// Reset timer when new question arrives
		const unsubscribe = onMessage('QUESTION', () => {
			questionStartTime = Date.now();
		});

		return () => {
			unsubscribe();
		};
	});

	function handleAnswer(answer: string) {
		if (!$currentQuestion || !$currentPlayerId) return;

		const timeTaken = Date.now() - questionStartTime;
		submitAnswer($currentPlayerId, $currentQuestion.id, answer, timeTaken);
		dispatch('answerSubmitted', {
			questionId: $currentQuestion.id,
			answer,
			timeTaken
		});
	}

	function formatTime(ms: number): string {
		return `${(ms / 1000).toFixed(1)}s`;
	}
</script>

{#if minimized}
	<!-- Minimized overlay -->
	<div class="fixed top-4 right-4 z-50 card-game p-3 max-w-xs">
		<div class="flex items-center justify-between mb-2">
			<h3 class="text-sm font-semibold text-text-light">
				{t('multiplayer.overlay.room', undefined, currentLocale) || 'Room'}
			</h3>
			{#if onExpand}
				<button
					on:click={onExpand}
					class="text-xs btn-secondary px-2 py-1"
				>
					{t('multiplayer.overlay.expand', undefined, currentLocale) || 'Expand'}
				</button>
			{/if}
		</div>

		<!-- Current Score -->
		{#if currentPlayer}
			<div class="text-xs text-text-muted mb-2">
				{t('multiplayer.overlay.your_score', undefined, currentLocale) || 'Your Score'}: 
				<span class="font-bold text-text-light">{currentPlayer.score}</span>
			</div>
		{/if}

		<!-- Leaderboard (top 3) -->
		<div class="space-y-1">
			{#each leaderboard.slice(0, 3) as player, index}
				<div class="flex items-center justify-between text-xs">
					<span class="text-text-muted">
						{index + 1}. {player.name}
					</span>
					<span class="font-semibold text-text-light">{player.score}</span>
				</div>
			{/each}
		</div>

		<!-- Question Progress -->
		{#if $currentRoom && $currentQuestion}
			<div class="mt-2 pt-2 border-t border-border">
				<div class="text-xs text-text-muted">
					{t('multiplayer.overlay.question', undefined, currentLocale) || 'Question'} 
					{$currentRoom.questionIndex + 1} / {$currentRoom.config.numQuestions}
				</div>
			</div>
		{/if}
	</div>
{:else}
	<!-- Full overlay -->
	<div class="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm flex items-center justify-center p-4">
		<div class="card-game max-w-2xl w-full max-h-[90vh] overflow-y-auto">
			<h2 class="text-2xl font-bold text-sandy-light mb-6">
				{t('multiplayer.overlay.leaderboard', undefined, currentLocale) || 'Leaderboard'}
			</h2>

			{#if $gameFinished}
				<!-- Game Finished -->
				<div class="mb-6 text-center">
					<h3 class="text-xl font-bold text-success mb-2">
						{t('multiplayer.overlay.game_finished', undefined, currentLocale) || 'Game Finished!'}
					</h3>
					{#if currentPlayer}
						<p class="text-text-light">
							{t('multiplayer.overlay.final_score', undefined, currentLocale) || 'Your Final Score'}: 
							<span class="font-bold text-2xl text-primary">{currentPlayer.score}</span>
						</p>
						<p class="text-text-muted mt-2">
							{t('multiplayer.overlay.rank', undefined, currentLocale) || 'Rank'}: 
							<span class="font-semibold">{playerRank} / {totalPlayers}</span>
						</p>
					{/if}
				</div>
			{/if}

			<!-- Full Leaderboard -->
			<div class="space-y-2">
				{#each leaderboard as player, index}
					<div 
						class="flex items-center justify-between p-3 rounded-lg {player.id === $currentPlayerId ? 'bg-primary/20' : 'bg-bg-secondary'}"
					>
						<div class="flex items-center gap-3">
							<span class="text-lg font-bold text-text-muted w-8">
								{index + 1}
							</span>
							<span class="text-text-light font-medium">
								{player.name}
								{#if player.id === $currentPlayerId}
									<span class="text-primary ml-2">(You)</span>
								{/if}
							</span>
						</div>
						<div class="flex items-center gap-4">
							{#if player.streak > 0}
								<span class="text-xs text-text-muted">
									🔥 {player.streak}
								</span>
							{/if}
							<span class="text-lg font-bold text-sandy-light">
								{player.score}
							</span>
						</div>
					</div>
				{/each}
			</div>
		</div>
	</div>
{/if}
