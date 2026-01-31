<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { RoomSetup, RoomLobby, GameView } from '$lib/components/multiplayer';
	import { createRoom, getRoomByCode } from '$lib/api/multiplayer';
	import { currentRoom, currentPlayerId, currentQuestion, resetMultiplayerState, onMessage, wsError } from '$lib/stores/multiplayer';
	import type { RoomConfig } from '$lib/stores/multiplayer';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';

	type View = 'setup' | 'lobby' | 'join' | 'game';

	let view: View = 'setup';
	let isLoading = false;
	let error: string | null = null;
	let playerName = '';
	let roomCode = '';
	let roomPassword = '';
	let selectedGameMode: string = 'flag';
	let roomRequiresPassword = false;

	// Listen for game start (QUESTION message)
	let unsubscribeQuestion: (() => void) | null = null;

	onMount(() => {
		const code = $page.url.searchParams.get('code');
		const mode = $page.url.searchParams.get('mode');
		if (code) {
			roomCode = code;
			view = 'join';
			handleJoinRoom();
		}
		if (mode) {
			selectedGameMode = mode;
		}

		// Listen for QUESTION messages to switch to game view
		unsubscribeQuestion = onMessage('QUESTION', () => {
			if (view === 'lobby') {
				view = 'game';
			}
		});

		// Also check if room status changes to playing
		const unsubscribeRoomStateGame = onMessage('ROOM_STATE', (msg) => {
			if (msg.type === 'ROOM_STATE' && msg.room.status === 'playing' && view === 'lobby') {
				view = 'game';
			}
		});

		return () => {
			if (unsubscribeQuestion) unsubscribeQuestion();
			unsubscribeRoomStateGame();
		};
	});

	async function handleCreateRoom(event: CustomEvent<{ config: RoomConfig; playerName: string }>) {
		isLoading = true;
		error = null;

		try {
			const result = await createRoom({
				hostName: event.detail.playerName,
				config: event.detail.config
			});

			playerName = event.detail.playerName;
			currentPlayerId.set(result.hostId);
			currentRoom.set(result.room);
			view = 'lobby';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create room';
		} finally {
			isLoading = false;
		}
	}

	async function handleJoinRoom() {
		if (!roomCode.trim()) {
			error = 'Room code is required';
			return;
		}

		if (!playerName.trim()) {
			error = 'Player name is required';
			return;
		}

		isLoading = true;
		error = null;

		try {
			const room = await getRoomByCode(roomCode);
			
			// Check if room requires password (password field exists and is not empty)
			// Note: Backend may omit empty passwords, so if password field exists, it requires one
			roomRequiresPassword = room.config.password !== undefined && room.config.password !== '';
			
			// If room requires password but none provided, show error
			if (roomRequiresPassword && !roomPassword.trim()) {
				error = 'This room requires a password';
				isLoading = false;
				return;
			}

			// Generate player ID (in real app, might come from auth)
			const playerId = `player_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
			currentPlayerId.set(playerId);
			
			// Import WebSocket functions
			const { connectWebSocket, joinRoom: wsJoinRoom } = await import('$lib/stores/multiplayer');
			
			// Set up listeners BEFORE connecting
			let handlersActive = true;
			let joinedSuccessfully = false;
			
			const unsubscribeError = onMessage('ERROR', (msg) => {
				if (!handlersActive) return;
				
				if (msg.type === 'ERROR') {
					// If it's a password error, stay in join form and show error
					if (msg.code === 'multiplayer.error.invalid_password') {
						error = t(msg.code, undefined, $locale) || 'Invalid password';
						view = 'join'; // Stay in join form
						isLoading = false;
						handlersActive = false;
						unsubscribeError();
						unsubscribeRoomState();
						return;
					}
					// Other errors - show in join form if still in join view
					if (view === 'join') {
						error = t(msg.code, undefined, $locale) || msg.code;
						isLoading = false;
					}
				}
			});
			
			const unsubscribeRoomState = onMessage('ROOM_STATE', (msg) => {
				if (!handlersActive) return;
				
				if (msg.type === 'ROOM_STATE' && !joinedSuccessfully) {
					// Successfully joined - switch to lobby
					joinedSuccessfully = true;
					currentRoom.set(msg.room);
					view = 'lobby';
					isLoading = false;
					handlersActive = false;
					unsubscribeError();
					unsubscribeRoomState();
				}
			});
			
			// Connect and join via WebSocket
			try {
				await connectWebSocket(room.id, playerId);
				wsJoinRoom(room.id, room.code, playerId, playerName, roomPassword);
				
				// Set a timeout to clean up handlers if no response
				setTimeout(() => {
					if (handlersActive && !joinedSuccessfully) {
						// No response received - assume connection issue
						if (view === 'join') {
							error = 'Failed to join room. Please try again.';
							isLoading = false;
						}
						handlersActive = false;
						unsubscribeError();
						unsubscribeRoomState();
					}
				}, 5000); // 5 second timeout
			} catch (err) {
				error = err instanceof Error ? err.message : 'Failed to connect to room';
				isLoading = false;
				handlersActive = false;
				unsubscribeError();
				unsubscribeRoomState();
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to join room';
			// If error is about invalid password, keep password field visible
			if (err instanceof Error && err.message.includes('password')) {
				roomRequiresPassword = true;
			}
			isLoading = false;
		}
	}

	function handleLobbyStart() {
		// Game will start via WebSocket, QUESTION message will trigger view change
		// Don't change view here - wait for QUESTION message
	}

	function handleLobbyLeave() {
		resetMultiplayerState();
		view = 'setup';
		roomCode = '';
	}
</script>

<svelte:head>
	<title>Multiplayer Debug - Flagged It</title>
</svelte:head>

<div class="min-h-screen p-4 md:p-8">
	<div class="max-w-4xl mx-auto">
		<div class="mb-6">
			<h1 class="text-4xl md:text-5xl font-bold text-sandy-light mb-2">
				<span class="emoji-blue mr-2">🎮</span>
				<span class="gradient-text">Multiplayer Debug</span>
			</h1>
			<p class="text-lg text-text-muted">
				Test multiplayer room creation and joining. This is a temporary debug entry point.
			</p>
		</div>

		{#if view === 'setup'}
			<div class="flex flex-col items-center gap-4 mb-8">
				<div class="flex gap-4">
					<button
						on:click={() => view = 'join'}
						class="btn-secondary"
					>
						Join Room
					</button>
					<a href="/debug" class="btn-secondary">
						← Back to Debug
					</a>
				</div>
			</div>
			<RoomSetup
				{isLoading}
				{error}
				bind:playerName
				on:create={handleCreateRoom}
			/>
		{:else if view === 'join'}
			<div class="card-game max-w-md mx-auto">
				<h2 class="text-2xl font-bold text-sandy-light mb-6 text-center">
					Join Room
				</h2>
			<form on:submit|preventDefault={handleJoinRoom} class="space-y-4">
				<div>
					<label for="join-player-name" class="block text-sm font-medium text-text-muted mb-2">
						Your Name
					</label>
					<input
						id="join-player-name"
						type="text"
						bind:value={playerName}
						placeholder="Enter your name"
						class="w-full px-4 py-2 bg-bg-secondary border border-border rounded-lg text-text-light"
						maxlength="20"
					/>
				</div>
				<div>
					<label for="join-room-code" class="block text-sm font-medium text-text-muted mb-2">
						Room Code
					</label>
					<input
						id="join-room-code"
						type="text"
						bind:value={roomCode}
						placeholder="Enter room code"
						class="w-full px-4 py-2 bg-bg-secondary border border-border rounded-lg text-text-light text-center text-2xl font-bold tracking-wider uppercase"
						maxlength="6"
						on:input={() => {
							// Reset password requirement when room code changes
							roomRequiresPassword = false;
							roomPassword = '';
						}}
					/>
				</div>
				<div>
					<label for="join-room-password" class="block text-sm font-medium text-text-muted mb-2">
						Password {roomRequiresPassword ? '(Required)' : '(Optional - if room has one)'}
					</label>
					<input
						id="join-room-password"
						type="password"
						bind:value={roomPassword}
						placeholder="Enter room password (if required)"
						class="w-full px-4 py-2 bg-bg-secondary border border-border rounded-lg text-text-light"
						maxlength="50"
					/>
					{#if roomRequiresPassword}
						<p class="text-xs text-text-muted mt-1">This room requires a password to join</p>
					{/if}
				</div>
					{#if error}
						<div class="p-4 bg-error/20 border border-error rounded-lg">
							<p class="text-error font-semibold text-center">{error}</p>
						</div>
					{/if}
					<div class="flex gap-4">
						<button
							type="button"
							on:click={() => { view = 'setup'; roomCode = ''; }}
							class="btn-secondary flex-1"
						>
							Back
						</button>
						<button
							type="submit"
							disabled={isLoading || !playerName.trim() || !roomCode.trim() || (roomRequiresPassword && !roomPassword.trim())}
							class="btn-primary flex-1"
						>
							{isLoading ? 'Joining...' : 'Join'}
						</button>
					</div>
				</form>
			</div>
		{:else if view === 'lobby' && $currentRoom}
			<div class="mb-4">
				<a href="/debug" class="btn-secondary">
					← Back to Debug
				</a>
			</div>
			{#key $currentRoom?.id}
				<RoomLobby
					room={$currentRoom}
					playerId={$currentPlayerId || ''}
					{playerName}
					password={roomPassword}
					on:start={handleLobbyStart}
					on:leave={handleLobbyLeave}
				/>
			{/key}
		{:else if view === 'game' && $currentRoom}
			<div class="mb-4">
				<a href="/debug" class="btn-secondary">
					← Back to Debug
				</a>
			</div>
			<GameView />
		{/if}
	</div>
</div>
