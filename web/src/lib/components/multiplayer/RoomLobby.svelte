<script lang="ts">
	import { createEventDispatcher, onMount, onDestroy } from 'svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import {
		currentRoom,
		currentPlayers,
		currentPlayerId,
		isHost,
		wsConnected,
		wsError,
		connectWebSocket,
		joinRoom,
		setReady,
		startGame,
		onMessage,
		resetMultiplayerState,
		disconnectWebSocket
	} from '$lib/stores/multiplayer';
	import type { Room, Player } from '$lib/stores/multiplayer';

	export let room: Room;
	export let playerId: string;
	export let playerName: string;
	export let password: string = '';

	$: currentLocale = $locale;
	// Use currentRoom from store for reactive updates, fallback to prop
	$: displayRoom = $currentRoom || room;
	$: roomCode = displayRoom?.code || '';
	$: shareUrl = typeof window !== 'undefined' ? `${window.location.origin}/multiplayer/${roomCode}` : '';
	$: players = Object.values($currentPlayers);
	$: allReady = players.length > 0 && players.every(p => p.isReady);
	$: currentPlayer = $currentPlayers[playerId];
	$: playerIsReady = currentPlayer?.isReady || false;
	$: host = $isHost;

	let copied = false;
	let connecting = false;

	let mounted = false;

	onMount(() => {
		if (!room || !playerId) {
			return;
		}

		// Prevent double mounting
		if (mounted) {
			return;
		}
		mounted = true;

		// Connect WebSocket if not already connected
		if (!$wsConnected) {
			connecting = true;
			connectWebSocket(room.id, playerId)
				.then(() => {
					// Join room via WebSocket
					joinRoom(room.id, room.code, playerId, playerName, password);
				})
				.catch((err) => {
					console.error('Failed to connect:', err);
				})
				.finally(() => {
					connecting = false;
				});
		}

		// Listen for room state updates
		// Don't mutate the prop - the store will update automatically
		const unsubscribe = onMessage('ROOM_STATE', () => {
			// Room state is already updated in the store, no need to mutate prop
		});

		return () => {
			mounted = false;
			unsubscribe();
		};
	});

	onDestroy(() => {
		// Don't disconnect WebSocket here - it should persist when transitioning to game view
		// WebSocket will be disconnected in handleLeave() when user explicitly leaves
		// or when the parent component resets state
	});

	function handleReady() {
		setReady(playerId, !playerIsReady);
	}

	const dispatch = createEventDispatcher<{
		start: void;
		leave: void;
	}>();

	function handleStart() {
		// Check if there are at least 2 players
		if (players.length < 2) {
			return;
		}
		
		if (host && allReady) {
			startGame(playerId);
			dispatch('start');
		}
	}

	function handleLeave() {
		// Disconnect WebSocket first so backend knows player left
		disconnectWebSocket();
		// Don't reset state here - let parent handle it via the 'leave' event
		// This prevents double cleanup
		dispatch('leave');
	}

	function copyShareUrl() {
		if (typeof navigator !== 'undefined' && navigator.clipboard) {
			navigator.clipboard.writeText(shareUrl);
			copied = true;
			setTimeout(() => copied = false, 2000);
		}
	}
</script>

<div class="card-game max-w-2xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<h2 class="text-2xl font-bold text-sandy-light">
			{t('multiplayer.lobby.title', undefined, currentLocale) || 'Game Lobby'}
		</h2>
		<button
			on:click={handleLeave}
			class="btn-secondary text-sm"
		>
			{t('multiplayer.lobby.leave', undefined, currentLocale) || 'Leave'}
		</button>
	</div>

	{#if connecting}
		<div class="flex items-center justify-center py-8">
			<LoadingSpinner />
			<span class="ml-4 text-text-muted">
				{t('multiplayer.lobby.connecting', undefined, currentLocale) || 'Connecting...'}
			</span>
		</div>
	{:else if $wsError}
		<div class="p-4 bg-error/20 border border-error rounded-lg mb-4">
			<p class="text-error font-semibold text-center">
				{t($wsError, undefined, currentLocale) || $wsError}
			</p>
		</div>
	{:else}
		<!-- Room Code -->
		<div class="mb-6">
			<div class="block text-sm font-medium text-text-muted mb-2">
				{t('multiplayer.lobby.room_code', undefined, currentLocale) || 'Room Code'}
			</div>
			<div class="flex items-center gap-2">
					<div class="flex-1 px-4 py-3 bg-bg-secondary border border-border rounded-lg text-2xl font-bold text-center text-sandy-light tracking-wider">
					{displayRoom?.code || roomCode}
				</div>
				<button
					on:click={copyShareUrl}
					class="btn-secondary px-4 py-3"
					title={t('multiplayer.lobby.copy_url', undefined, currentLocale) || 'Copy share URL'}
					aria-label={t('multiplayer.lobby.copy_url', undefined, currentLocale) || 'Copy share URL'}
				>
					{copied ? '✓' : '📋'}
				</button>
			</div>
			<p class="text-xs text-text-muted mt-2 text-center">
				{t('multiplayer.lobby.share_code', undefined, currentLocale) || 'Share this code with friends to join'}
			</p>
		</div>

		<!-- Players List -->
		<div class="mb-6">
			<h3 class="text-lg font-semibold text-text-light mb-3" id="players-list">
				{t('multiplayer.lobby.players', undefined, currentLocale) || 'Players'} ({players.length}/{displayRoom?.config.maxPlayers || room.config.maxPlayers})
			</h3>
			<div class="space-y-2">
				{#each players as player (player.id)}
					<div class="flex items-center justify-between p-3 bg-bg-secondary rounded-lg">
						<div class="flex items-center gap-3">
							<span class="text-text-light font-medium">
								{player.name}
								{#if player.isHost}
									<span class="text-primary ml-2">👑</span>
								{/if}
							</span>
						</div>
						<div class="flex items-center gap-2">
							{#if player.isReady}
								<span class="text-success">✓</span>
							{:else}
								<span class="text-text-muted">○</span>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</div>

		<!-- Room Settings (Host only) -->
		{#if host}
			<div class="mb-6 p-4 bg-bg-secondary rounded-lg">
				<h3 class="text-sm font-semibold text-text-muted mb-2">
					{t('multiplayer.lobby.room_settings', undefined, currentLocale) || 'Room Settings'}
				</h3>
				<div class="text-sm text-text-light space-y-1">
					<p>
						{t('multiplayer.lobby.game_mode', undefined, currentLocale) || 'Game'}: 
						<span class="font-medium">{displayRoom?.config.gameMode || room.config.gameMode}</span>
					</p>
					<p>
						{t('multiplayer.lobby.questions', undefined, currentLocale) || 'Questions'}: 
						<span class="font-medium">{displayRoom?.config.numQuestions || room.config.numQuestions}</span>
					</p>
					<p>
						{t('multiplayer.lobby.difficulty', undefined, currentLocale) || 'Difficulty'}: 
						<span class="font-medium">{displayRoom?.config.difficulty || room.config.difficulty}</span>
					</p>
				</div>
			</div>
		{/if}

		<!-- Ready Button -->
		<div class="mb-4">
			<button
				on:click={handleReady}
				class="w-full {playerIsReady ? 'btn-secondary' : 'btn-primary'}"
			>
				{playerIsReady
					? (t('multiplayer.lobby.not_ready', undefined, currentLocale) || 'Not Ready')
					: (t('multiplayer.lobby.ready', undefined, currentLocale) || 'Ready')
				}
			</button>
		</div>

		<!-- Start Button (Host only) -->
		{#if host}
			<button
				on:click={handleStart}
				disabled={!allReady || players.length < 2}
				class="btn-primary w-full disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{#if players.length < 2}
					{t('multiplayer.lobby.need_more_players', undefined, currentLocale) || 'Need at least 2 players to start'}
				{:else if allReady}
					{t('multiplayer.lobby.start_game', undefined, currentLocale) || 'Start Game'}
				{:else}
					{t('multiplayer.lobby.waiting_players', undefined, currentLocale) || 'Waiting for all players...'}
				{/if}
			</button>
		{/if}
	{/if}
</div>
