<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import RegionDropdown from '$lib/components/ui/RegionDropdown.svelte';
	import type { RoomConfig, GameMode, DifficultyLevel } from '$lib/stores/multiplayer';

	export let isLoading: boolean = false;
	export let error: string | null = null;
	export let playerName: string = '';

	const dispatch = createEventDispatcher<{
		create: { config: RoomConfig; playerName: string };
	}>();

	$: currentLocale = $locale;

	// Default config
	let config: RoomConfig = {
		gameMode: 'flag',
		numQuestions: 10,
		difficulty: 'medium',
		timeLimit: 0, // unlimited
		maxPlayers: 10,
		region: '',
		categories: [],
		isPublic: false,
		password: ''
	};

	let showPassword = false;
	let usePassword = false;

	// Game mode options
	$: gameModes = [
		{ value: 'flag', label: t('game.flag.title', undefined, currentLocale) },
		{ value: 'shape', label: t('game.shape.title', undefined, currentLocale) },
		{ value: 'capital', label: t('game.capital.title', undefined, currentLocale) },
		{ value: 'higherlower', label: t('game.higher_lower.title', undefined, currentLocale) },
		{ value: 'facts', label: t('game.facts.title', undefined, currentLocale) }
	];

	// Difficulty options
	$: difficulties = [
		{ value: 'easy', label: t('game.flag.difficulty.regular', undefined, currentLocale) },
		{ value: 'medium', label: t('game.flag.difficulty.expert', undefined, currentLocale) },
		{ value: 'hard', label: 'Hard' }
	];

	// Region options
	$: regions = [
		{ value: '', label: t('region.world', undefined, currentLocale) },
		{ value: 'Africa', label: t('region.africa', undefined, currentLocale) },
		{ value: 'Americas', label: t('region.americas', undefined, currentLocale) },
		{ value: 'Asia', label: t('region.asia', undefined, currentLocale) },
		{ value: 'Europe', label: t('region.europe', undefined, currentLocale) },
		{ value: 'Oceania', label: t('region.oceania', undefined, currentLocale) }
	];

	function handleCreate() {
		if (!playerName.trim()) {
			error = 'Player name is required';
			return;
		}

		const finalConfig = {
			...config,
			password: usePassword ? config.password : ''
		};

		dispatch('create', { config: finalConfig, playerName: playerName.trim() });
	}

	function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		handleCreate();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' && !isLoading && playerName.trim()) {
			event.preventDefault();
			handleCreate();
		}
	}
</script>

<div class="card-game max-w-2xl mx-auto">
	<h2 class="text-2xl font-bold text-sandy-light mb-6 text-center">
		{t('multiplayer.setup.title', undefined, currentLocale) || 'Create Game Room'}
	</h2>

	<form on:submit={handleSubmit} class="space-y-6">
		<!-- Player Name -->
		<div>
			<label for="player-name" class="block text-sm font-medium text-text-muted mb-2">
				{t('multiplayer.setup.player_name', undefined, currentLocale) || 'Your Name'}
			</label>
			<input
				id="player-name"
				type="text"
				bind:value={playerName}
				placeholder={t('multiplayer.setup.player_name_placeholder', undefined, currentLocale) || 'Enter your name'}
				class="w-full px-4 py-2 bg-bg-secondary border border-border rounded-lg text-text-light focus:outline-none focus:ring-2 focus:ring-primary"
				maxlength="20"
				on:keydown={handleKeydown}
			/>
		</div>

		<!-- Game Mode -->
		<div>
			<label for="game-mode" class="block text-sm font-medium text-text-muted mb-2">
				{t('multiplayer.setup.game_mode', undefined, currentLocale) || 'Game Mode'}
			</label>
			<select
				id="game-mode"
				bind:value={config.gameMode}
				class="w-full px-4 py-2 bg-bg-secondary border border-border rounded-lg text-text-light focus:outline-none focus:ring-2 focus:ring-primary"
			>
				{#each gameModes as mode}
					<option value={mode.value}>{mode.label}</option>
				{/each}
			</select>
		</div>

		<!-- Number of Questions -->
		<div>
			<label for="num-questions" class="block text-sm font-medium text-text-muted mb-2">
				{t('multiplayer.setup.num_questions', undefined, currentLocale) || 'Number of Questions'}
			</label>
			<input
				id="num-questions"
				type="number"
				bind:value={config.numQuestions}
				min="5"
				max="50"
				step="5"
				class="w-full px-4 py-2 bg-bg-secondary border border-border rounded-lg text-text-light focus:outline-none focus:ring-2 focus:ring-primary"
			/>
		</div>

		<!-- Difficulty -->
		<div>
			<label for="difficulty" class="block text-sm font-medium text-text-muted mb-2">
				{t('multiplayer.setup.difficulty', undefined, currentLocale) || 'Difficulty'}
			</label>
			<select
				id="difficulty"
				bind:value={config.difficulty}
				class="w-full px-4 py-2 bg-bg-secondary border border-border rounded-lg text-text-light focus:outline-none focus:ring-2 focus:ring-primary"
			>
				{#each difficulties as diff}
					<option value={diff.value}>{diff.label}</option>
				{/each}
			</select>
		</div>

		<!-- Time Limit -->
		<div>
			<label for="time-limit" class="block text-sm font-medium text-text-muted mb-2">
				{t('multiplayer.setup.time_limit', undefined, currentLocale) || 'Time Limit (seconds)'}
			</label>
			<input
				id="time-limit"
				type="number"
				bind:value={config.timeLimit}
				min="0"
				max="60"
				step="5"
				class="w-full px-4 py-2 bg-bg-secondary border border-border rounded-lg text-text-light focus:outline-none focus:ring-2 focus:ring-primary"
			/>
			<p class="text-xs text-text-muted mt-1">
				{t('multiplayer.setup.time_limit_hint', undefined, currentLocale) || '0 = unlimited'}
			</p>
		</div>

		<!-- Max Players -->
		<div>
			<label for="max-players" class="block text-sm font-medium text-text-muted mb-2">
				{t('multiplayer.setup.max_players', undefined, currentLocale) || 'Max Players'}
			</label>
			<input
				id="max-players"
				type="number"
				bind:value={config.maxPlayers}
				min="2"
				max="20"
				class="w-full px-4 py-2 bg-bg-secondary border border-border rounded-lg text-text-light focus:outline-none focus:ring-2 focus:ring-primary"
			/>
		</div>

		<!-- Region -->
		<div>
			<RegionDropdown
				regions={regions}
				bind:selected={config.region}
				label={t('game.setup.select_region', undefined, currentLocale)}
			/>
		</div>

		<!-- Public/Private -->
		<div class="flex items-center gap-4">
			<label class="flex items-center gap-2 cursor-pointer">
				<input
					type="checkbox"
					bind:checked={config.isPublic}
					class="w-4 h-4 text-primary bg-bg-secondary border-border rounded focus:ring-primary"
				/>
				<span class="text-text-light">
					{t('multiplayer.setup.public_room', undefined, currentLocale) || 'Public Room'}
				</span>
			</label>
		</div>

		<!-- Password Protection -->
		<div class="flex items-center gap-4">
			<label class="flex items-center gap-2 cursor-pointer">
				<input
					type="checkbox"
					bind:checked={usePassword}
					class="w-4 h-4 text-primary bg-bg-secondary border-border rounded focus:ring-primary"
				/>
				<span class="text-text-light">
					{t('multiplayer.setup.password_protect', undefined, currentLocale) || 'Password Protect'}
				</span>
			</label>
		</div>

		{#if usePassword}
			<div>
				<label for="password" class="block text-sm font-medium text-text-muted mb-2">
					{t('multiplayer.setup.password', undefined, currentLocale) || 'Password'}
				</label>
				<input
					id="password"
					type="password"
					bind:value={config.password}
					placeholder={t('multiplayer.setup.password_placeholder', undefined, currentLocale) || 'Enter password'}
					class="w-full px-4 py-2 bg-bg-secondary border border-border rounded-lg text-text-light focus:outline-none focus:ring-2 focus:ring-primary"
					on:keydown={handleKeydown}
				/>
			</div>
		{/if}

		{#if error}
			<div class="p-4 bg-error/20 border border-error rounded-lg">
				<p class="text-error font-semibold text-center">{error}</p>
			</div>
		{/if}

		<button
			type="submit"
			disabled={isLoading}
			class="btn-primary w-full"
		>
			{isLoading 
				? (t('multiplayer.setup.creating', undefined, currentLocale) || 'Creating...')
				: (t('multiplayer.setup.create_room', undefined, currentLocale) || 'Create Room')
			}
		</button>
	</form>
</div>
