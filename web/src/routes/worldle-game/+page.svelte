<script lang="ts">
	import { onMount } from 'svelte';
	import { 
		GameSetupScreen, 
		GameOverScreen 
	} from '$lib/components/game';
	import { browser } from '$app/environment';
	import { startWorldleGame, submitGuess, formatNumber, getGameState, type GuessEntry } from '$lib/api/worldleGame';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import { getCountryNameForLocale } from '$lib/utils/countryNames';
	import { getAllCountries } from '$lib/api/data';
	import type { Country } from '$lib/types';

	let sessionId: string | null = null;
	let guesses: GuessEntry[] = [];
	let guessCount = 0;
	let isLoading = false;
	let error: string | null = null;
	let gameStarted = false;
	let gameComplete = false;
	let guessInput = '';
	let allCountries: Country[] = [];
	let countriesLoaded = false;
	let guessInputElement: HTMLInputElement | null = null;

	// Reactive translations
	$: currentLocale = $locale;
	$: worldleTitle = t('game.worldle.title', undefined, currentLocale) || 'Worldle';
	$: worldleDescription = t('game.worldle.description', undefined, currentLocale) || 'Guess the country by comparing attributes!';
	$: makeGuessText = t('game.guessing.make_guess', undefined, currentLocale);
	$: enterCountryText = t('game.guessing.enter_country', undefined, currentLocale);
	$: guessText = t('game.guessing.guess', undefined, currentLocale);
	$: historyText = t('game.guessing.history', undefined, currentLocale);
	$: flagText = t('game.guessing.flag', undefined, currentLocale);
	$: countryText = t('game.guessing.country', undefined, currentLocale);
	$: continentText = t('game.guessing.continent', undefined, currentLocale);
	$: populationText = t('game.guessing.population', undefined, currentLocale);
	$: areaText = t('game.guessing.area', undefined, currentLocale);
	$: notFoundText = t('game.guessing.not_found', undefined, currentLocale);
	$: correctText = t('game.guessing.correct', undefined, currentLocale);
	$: playAgainText = t('game.over.play_again', undefined, currentLocale);
	$: excellentMessage = t('game.over.excellent', undefined, currentLocale);

	// Re-fetch game state when locale changes (to update country names in guesses)
	let previousLocale: string = currentLocale;
	$: if (gameStarted && sessionId && currentLocale !== previousLocale && !isLoading && !gameComplete) {
		previousLocale = currentLocale;
		// Re-fetch the game state with new locale to get translated country names
		getGameState(sessionId).then(state => {
			guesses = state.guesses || guesses;
			guessCount = state.guessCount || guessCount;
			guessInput = '';
		}).catch(err => {
			console.error('Failed to reload game state with new locale:', err);
		});
	}

	onMount(async () => {
		// Only load countries in browser, not during SSR
		if (!browser) {
			// Auto-start game immediately if no game mode selection is needed
			if (!gameStarted && !gameComplete) {
				handleStartGame();
			}
			return;
		}
		
		try {
			const result = await getAllCountries();
			allCountries = result.countries;
			countriesLoaded = true;
		} catch (err) {
			// Silently fail - will use English names as fallback
			console.warn('Failed to load countries for translation (API server may not be running):', err);
		}
		
		// Auto-start game if no game mode selection is needed
		if (!gameStarted && !gameComplete) {
			handleStartGame();
		}
	});

	async function handleStartGame(event?: CustomEvent<{ region?: string }>) {
		isLoading = true;
		error = null;
		gameComplete = false;
		guesses = [];
		guessCount = 0;
		guessInput = '';
		
		try {
			const result = await startWorldleGame();
			sessionId = result.sessionId;
			guessCount = result.guessCount;
			gameStarted = true;
			gameComplete = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to start game';
			console.error('Start game error:', err);
		} finally {
			isLoading = false;
		}
	}

	async function handleSubmitGuess() {
		if (!sessionId || !guessInput.trim() || isLoading) return;

		const countryName = guessInput.trim();
		isLoading = true;
		error = null;

		try {
			const result = await submitGuess(sessionId, countryName);
			
			if (!result.isValidGuess) {
				error = notFoundText || result.error || null;
				isLoading = false;
				// Focus input so user can correct their guess
				setTimeout(() => {
					guessInputElement?.focus();
				}, 0);
				return;
			}

			if (result.guessEntry) {
				guesses = [...guesses, result.guessEntry];
				guessCount = result.guessCount;
			}

			if (result.isCorrect) {
				gameComplete = true;
				gameStarted = false;
			}

			guessInput = '';
			
			// Focus input after successful guess (if game is still active)
			if (!result.isCorrect && !gameComplete && guessInputElement) {
				// Use setTimeout to ensure DOM has updated
				setTimeout(() => {
					guessInputElement?.focus();
				}, 0);
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to submit guess';
			console.error('Submit guess error:', err);
		} finally {
			isLoading = false;
		}
	}

	function handleKeyPress(event: KeyboardEvent) {
		if (event.key === 'Enter' && !isLoading) {
			handleSubmitGuess();
		}
	}

	function handlePlayAgain() {
		gameStarted = false;
		gameComplete = false;
		guesses = [];
		guessCount = 0;
		sessionId = null;
		guessInput = '';
		error = null;
	}

	function getContinentClass(guess: GuessEntry): string {
		return guess.continentCorrect ? 'bg-success/30 text-success border-success' : 'bg-error/30 text-error border-error';
	}

	function getPopulationClass(guess: GuessEntry): string {
		if (guess.population.direction === 'correct') return 'bg-success text-white';
		if (guess.population.proximity === 'very_close') return 'bg-yellow-500/30 text-yellow-300 border-yellow-500';
		if (guess.population.proximity === 'close') return 'bg-orange-500/30 text-orange-300 border-orange-500';
		return 'bg-error/30 text-error border-error';
	}

	function getAreaClass(guess: GuessEntry): string {
		if (guess.area.direction === 'correct') return 'bg-success text-white';
		if (guess.area.proximity === 'very_close') return 'bg-yellow-500/30 text-yellow-300 border-yellow-500';
		if (guess.area.proximity === 'close') return 'bg-orange-500/30 text-orange-300 border-orange-500';
		return 'bg-error/30 text-error border-error';
	}

	function getPopulationArrow(guess: GuessEntry): string {
		if (guess.population.direction === 'correct') return '✓';
		// If guess is higher than target, target is lower → show ▼
		// If guess is lower than target, target is higher → show ▲
		return guess.population.direction === 'higher' ? '▼' : '▲';
	}

	function getAreaArrow(guess: GuessEntry): string {
		if (guess.area.direction === 'correct') return '✓';
		// If guess is higher than target, target is lower → show ▼
		// If guess is lower than target, target is higher → show ▲
		return guess.area.direction === 'higher' ? '▼' : '▲';
	}

	function getCountryNameForGuess(guess: GuessEntry): string {
		const country = allCountries.find(c => c.cca2 === guess.country.cca2);
		return country ? getCountryNameForLocale(country) : guess.country.name;
	}
</script>

<svelte:head>
	<title>{worldleTitle} - Flagged It</title>
</svelte:head>

<div class="min-h-screen p-4 md:p-8">
	<div class="max-w-6xl mx-auto">
		{#if !gameStarted && !gameComplete}
			<GameSetupScreen
				title={worldleTitle}
				emoji="🌍"
				description={worldleDescription}
				{isLoading}
				{error}
				showRegionSelector={false}
				startButtonText={t('game.setup.start_game', undefined, currentLocale)}
				on:start={handleStartGame}
			/>
		{:else if gameComplete}
			{@const lastGuess = guesses[guesses.length - 1]}
			{@const correctCountryName = lastGuess ? getCountryNameForGuess(lastGuess) : ''}
			<GameOverScreen
				title={correctText.replace('%s', correctCountryName)}
				score={1}
				totalRounds={1}
				excellentMessage={excellentMessage}
				on:playAgain={handlePlayAgain}
			/>
		{:else}
			<div class="space-y-6">
				<h1 class="text-4xl md:text-5xl font-bold text-center mb-6">
					<span class="gradient-text">{worldleTitle}</span>
				</h1>

				<!-- Guess Input -->
				<div class="card-game max-w-2xl mx-auto">
					<p class="text-lg text-text-muted mb-4 text-center">{makeGuessText}</p>
					<div class="flex flex-col sm:flex-row gap-4">
						<input
							type="text"
							bind:this={guessInputElement}
							bind:value={guessInput}
							on:keypress={handleKeyPress}
							placeholder={enterCountryText}
							disabled={isLoading || gameComplete}
							class="flex-1 px-4 py-3 rounded-lg border-2 border-white/20 bg-white/5 text-sandy-light placeholder:text-text-muted focus:outline-none focus:border-primary disabled:opacity-50 min-w-0"
						/>
						<button
							on:click={handleSubmitGuess}
							disabled={isLoading || !guessInput.trim() || gameComplete}
							class="btn-primary px-6 sm:px-8 py-3 disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap flex-shrink-0"
						>
							{guessText}
						</button>
					</div>
					{#if error}
						<p class="text-error mt-4 text-center">{error}</p>
					{/if}
				</div>

				<!-- Guess History Table -->
				{#if guesses.length > 0}
					<div class="card-game overflow-x-auto">
						<h2 class="text-2xl font-bold mb-4">{historyText}</h2>
						<table class="w-full border-collapse">
							<thead>
								<tr class="border-b-2 border-white/20">
									<th class="px-4 py-3 text-left font-semibold">{flagText}</th>
									<th class="px-4 py-3 text-left font-semibold">{countryText}</th>
									<th class="px-4 py-3 text-left font-semibold">{continentText}</th>
									<th class="px-4 py-3 text-left font-semibold">{populationText}</th>
									<th class="px-4 py-3 text-left font-semibold">{areaText}</th>
								</tr>
							</thead>
							<tbody>
								{#each guesses.slice().reverse() as guess (guess.country.cca2)}
									<tr class="border-b border-white/10 hover:bg-white/5">
										<td class="px-4 py-3">
											<img 
												src={guess.country.flagUrl} 
												alt={guess.country.name}
												class="w-12 h-8 object-cover rounded"
											/>
										</td>
										<td class="px-4 py-3 font-semibold">{getCountryNameForGuess(guess)}</td>
										<td class="px-4 py-3">
											<span class="px-3 py-1 rounded border-2 {getContinentClass(guess)}">
												{guess.continent}
											</span>
										</td>
										<td class="px-4 py-3">
											<span class="px-3 py-1 rounded border-2 flex items-center gap-2 {getPopulationClass(guess)}">
												<span>{getPopulationArrow(guess)}</span>
												<span>{formatNumber(guess.country.population)}</span>
											</span>
										</td>
										<td class="px-4 py-3">
											<span class="px-3 py-1 rounded border-2 flex items-center gap-2 {getAreaClass(guess)}">
												<span>{getAreaArrow(guess)}</span>
												<span>{formatNumber(guess.country.area)} km²</span>
											</span>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>

<style>
	/* Light mode: table headers and borders */
	:global(:root.light) .card-game table thead tr {
		border-bottom-color: rgba(15, 23, 42, 0.2) !important;
	}

	:global(:root.light) .card-game table thead th {
		color: #0F172A !important;
	}

	:global(:root.light) .card-game table tbody tr {
		border-bottom-color: rgba(15, 23, 42, 0.1) !important;
	}

	:global(:root.light) .card-game table tbody tr:hover {
		background-color: rgba(15, 23, 42, 0.05) !important;
	}

	:global(:root.light) .card-game table tbody td {
		color: #0F172A !important;
	}

	/* Light mode: continent badges */
	:global(:root.light) .card-game .bg-success\/30 {
		background-color: rgba(16, 185, 129, 0.15) !important;
		color: #059669 !important;
		border-color: #10B981 !important;
	}

	:global(:root.light) .card-game .bg-error\/30 {
		background-color: rgba(239, 68, 68, 0.15) !important;
		color: #DC2626 !important;
		border-color: #EF4444 !important;
	}

	/* Light mode: population/area badges */
	:global(:root.light) .card-game .bg-yellow-500\/30 {
		background-color: rgba(234, 179, 8, 0.2) !important;
		color: #A16207 !important;
		border-color: #EAB308 !important;
	}

	:global(:root.light) .card-game .text-yellow-300 {
		color: #A16207 !important;
	}

	:global(:root.light) .card-game .bg-orange-500\/30 {
		background-color: rgba(249, 115, 22, 0.2) !important;
		color: #C2410C !important;
		border-color: #F97316 !important;
	}

	:global(:root.light) .card-game .text-orange-300 {
		color: #C2410C !important;
	}

	:global(:root.light) .card-game .text-error {
		color: #DC2626 !important;
	}

	:global(:root.light) .card-game .text-success {
		color: #059669 !important;
	}

	/* Light mode: table heading */
	:global(:root.light) .card-game h2 {
		color: #0F172A !important;
	}
</style>
