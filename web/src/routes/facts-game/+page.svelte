<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { 
		GameSetupScreen, 
		GameHeader, 
		GameOverScreen 
	} from '$lib/components/game';
	import { startFactsGame, submitGuess, nextRound, type GuessHistoryEntry } from '$lib/api/factsGame';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import { getCountryName } from '$lib/utils/countryNames';
	import { getAllCountries } from '$lib/api/debug';
	import type { Country } from '$lib/types';

	let sessionId: string | null = null;
	let currentFact: string = '';
	let factNumber: number = 0;
	let triesLeft: number = 3;
	let score = 0;
	let total = 0;
	let isLoading = false;
	let error: string | null = null;
	let gameStarted = false;
	let gameFinished = false;
	let guessInput = '';
	let guessHistory: GuessHistoryEntry[] = [];
	let statusMessage: string = '';
	let showFeedback = false;
	let isCorrect: boolean | null = null;
	let correctCountryName: string = '';
	let correctCountryCca2: string = '';
	let allCountries: Country[] = [];
	let countriesLoaded = false;
	const totalRounds = 5;

	// Reactive translations
	$: currentLocale = $locale;
	$: factsTitle = t('game.facts.title', undefined, currentLocale);
	$: factsDescription = t('game.facts.description', undefined, currentLocale) || t('home.game.facts.description', undefined, currentLocale) || t('game.facts.guess_country', undefined, currentLocale);
	$: makeGuessText = t('game.facts.guess_country', undefined, currentLocale);
	$: enterCountryText = t('game.guessing.enter_country', undefined, currentLocale);
	$: guessText = t('game.guessing.guess', undefined, currentLocale);
	$: previousGuessesText = t('game.facts.previous_guesses', undefined, currentLocale);
	$: triesLeftText = t('game.facts.tries_left', undefined, currentLocale);
	$: correctText = t('game.facts.correct', undefined, currentLocale);
	$: gameOverText = t('game.facts.game_over', undefined, currentLocale);
	$: wrongNextText = t('game.facts.wrong_next', undefined, currentLocale);
	$: wrongNoMoreText = t('game.facts.wrong_no_more', undefined, currentLocale);
	$: excellentMessage = t('game.over.excellent', undefined, currentLocale);
	$: playAgainText = t('game.over.play_again', undefined, currentLocale);

	onMount(async () => {
		// Only load countries in browser, not during SSR
		if (!browser) {
			// Auto-start game immediately if no game mode selection is needed
			if (!gameStarted && !gameFinished) {
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
		if (!gameStarted && !gameFinished) {
			handleStartGame();
		}
	});

	async function handleStartGame(event?: CustomEvent<{ region?: string }>) {
		isLoading = true;
		error = null;
		gameFinished = false;
		score = 0;
		total = 0;
		guessHistory = [];
		statusMessage = '';
		showFeedback = false;
		guessInput = '';
		
		try {
			const result = await startFactsGame();
			sessionId = result.sessionId;
			currentFact = result.currentFact;
			factNumber = result.factNumber;
			triesLeft = result.triesLeft;
			score = result.score;
			total = result.total;
			gameStarted = true;
			gameFinished = false;
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
		showFeedback = false;

		try {
			const result = await submitGuess(sessionId, countryName);
			
			isCorrect = result.isCorrect;
			triesLeft = result.triesLeft;
			score = result.score;
			total = result.total;
			guessHistory = result.guessHistory;

			if (result.isCorrect) {
				// Correct guess - show success message
				if (result.correctCountry) {
					const country = allCountries.find(c => c.cca2 === result.correctCountry!.cca2);
					correctCountryName = country ? getCountryName(country, currentLocale) : result.correctCountry.name;
					correctCountryCca2 = result.correctCountry.cca2;
					statusMessage = correctText.replace('%s', correctCountryName);
				}
				showFeedback = true;
				isCorrect = true;

				// Wait 2 seconds, then start next round or finish game
				setTimeout(async () => {
					if (result.isComplete) {
						gameFinished = true;
						gameStarted = false;
					} else {
						await handleNextRound();
					}
				}, 2000);
			} else if (triesLeft === 0) {
				// No tries left - show game over
				if (result.correctCountry) {
					const country = allCountries.find(c => c.cca2 === result.correctCountry!.cca2);
					correctCountryName = country ? getCountryName(country, currentLocale) : result.correctCountry.name;
					correctCountryCca2 = result.correctCountry.cca2;
					statusMessage = gameOverText.replace('%s', correctCountryName).replace(' %s', '');
				}
				showFeedback = true;
				isCorrect = false;

				// Wait 2 seconds, then start next round or finish game
				setTimeout(async () => {
					if (result.isComplete) {
						gameFinished = true;
						gameStarted = false;
					} else {
						await handleNextRound();
					}
				}, 2000);
			} else {
				// Wrong guess, but tries left - show next fact
				if (result.nextFact) {
					currentFact = result.nextFact;
					factNumber = result.factNumber || factNumber + 1;
					statusMessage = wrongNextText;
				} else {
					statusMessage = wrongNoMoreText;
				}
				showFeedback = true;
				isCorrect = false;
				guessInput = '';
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to submit guess';
			console.error('Submit guess error:', err);
		} finally {
			isLoading = false;
		}
	}

	async function handleNextRound() {
		if (!sessionId) return;

		isLoading = true;
		error = null;
		guessHistory = [];
		statusMessage = '';
		showFeedback = false;
		guessInput = '';
		isCorrect = null;
		correctCountryName = '';

		try {
			const result = await nextRound(sessionId);
			currentFact = result.currentFact;
			factNumber = result.factNumber;
			triesLeft = result.triesLeft;
			score = result.score;
			total = result.total;
			
			if (result.isComplete) {
				gameFinished = true;
				gameStarted = false;
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to start next round';
			console.error('Next round error:', err);
		} finally {
			isLoading = false;
		}
	}

	function handleKeyPress(event: KeyboardEvent) {
		if (event.key === 'Enter' && !isLoading && guessInput.trim()) {
			handleSubmitGuess();
		}
	}

	function handlePlayAgain() {
		gameStarted = false;
		gameFinished = false;
		guessHistory = [];
		statusMessage = '';
		showFeedback = false;
		sessionId = null;
		guessInput = '';
		score = 0;
		total = 0;
		error = null;
	}

	function formatFact(fact: string): string {
		// Convert markdown-style bold **text** to HTML <strong>text</strong>
		return fact.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
	}

	function getGuessDisplayText(entry: GuessHistoryEntry): string {
		// Remove checkmark if present, clean the guess text
		return entry.guess.replace('✅', '').trim();
	}

	function isGuessCorrect(entry: GuessHistoryEntry, index: number): boolean {
		// Use isCorrect field from API if available
		if (entry.isCorrect !== undefined) {
			return entry.isCorrect;
		}
		// Fallback: check if it's the last entry and the result was correct
		return index === guessHistory.length - 1 && isCorrect === true;
	}
</script>

<svelte:head>
	<title>{factsTitle} - Flagged It</title>
</svelte:head>

<div class="min-h-screen p-4 md:p-8">
	<div class="max-w-4xl mx-auto">
		{#if !gameStarted && !gameFinished}
			<GameSetupScreen
				title={factsTitle}
				emoji="📚"
				description={factsDescription}
				{isLoading}
				{error}
				showRegionSelector={false}
				startButtonText={t('game.setup.start_game', undefined, currentLocale)}
				on:start={handleStartGame}
			/>
		{:else if gameFinished}
			<GameOverScreen
				title={t('game.over.title', undefined, currentLocale)}
				{score}
				totalRounds={totalRounds}
				excellentMessage={excellentMessage}
				on:playAgain={handlePlayAgain}
			/>
		{:else}
			<div class="space-y-6">
				<GameHeader
					{score}
					{total}
					currentRound={total + 1}
					{totalRounds}
				/>

				<!-- Current Fact Display -->
				{#if currentFact}
					<div class="card-game">
						<p class="text-lg text-text-light mb-4 prose prose-invert max-w-none" innerHTML={formatFact(currentFact)}></p>
					</div>
				{/if}

				<!-- Status Message and Feedback -->
				{#if statusMessage}
					<div class="card-game {isCorrect === true ? 'bg-success/20 border-success' : isCorrect === false ? 'bg-error/20 border-error' : 'bg-primary/20 border-primary'}">
						<p class="text-center font-semibold {isCorrect === true ? 'text-success' : isCorrect === false ? 'text-error' : 'text-primary'}">
							{statusMessage}
						</p>
						{#if triesLeft !== undefined && triesLeft > 0 && !isCorrect}
							<p class="text-center mt-2 text-text-muted">
								{triesLeftText.replace('%d', triesLeft.toString())}
							</p>
						{/if}
					</div>
				{/if}

				<!-- Guess Input -->
				{#if !showFeedback || (triesLeft > 0 && !isCorrect)}
					<div class="card-game">
						<p class="text-lg text-text-muted mb-4 text-center">{makeGuessText}</p>
						{#if triesLeft > 0 && !showFeedback}
							<p class="text-sm text-text-muted mb-2 text-center">{triesLeftText.replace('%d', triesLeft.toString())}</p>
						{/if}
						<div class="flex gap-4">
							<input
								type="text"
								bind:value={guessInput}
								on:keypress={handleKeyPress}
								placeholder={enterCountryText}
								disabled={isLoading || showFeedback || gameFinished}
								class="flex-1 px-4 py-3 rounded-lg border-2 border-white/20 bg-white/5 text-sandy-light placeholder:text-text-muted focus:outline-none focus:border-primary disabled:opacity-50"
							/>
							<button
								on:click={handleSubmitGuess}
								disabled={isLoading || !guessInput.trim() || showFeedback || gameFinished}
								class="btn-primary px-8 py-3 disabled:opacity-50 disabled:cursor-not-allowed"
							>
								{guessText}
							</button>
						</div>
						{#if error && !showFeedback}
							<p class="text-error mt-4 text-center">{error}</p>
						{/if}
					</div>
				{/if}

				<!-- Previous Guesses History -->
				{#if guessHistory.length > 0}
					<div class="card-game">
						<h2 class="text-2xl font-bold mb-4">{previousGuessesText}</h2>
						<div class="space-y-4">
							{#each guessHistory.slice().reverse() as entry, reverseIndex (entry.guess + entry.fact + String(reverseIndex))}
								{@const originalIndex = guessHistory.length - reverseIndex - 1}
								<div class="border-b border-white/10 pb-4 last:border-b-0 last:pb-0">
									<div class="flex items-start gap-3 mb-2">
										<span class="text-xl font-bold mt-1">
											{#if entry.isCorrect !== undefined && entry.isCorrect}
												<span class="text-success">✓</span>
											{:else}
												<span class="text-error">✗</span>
											{/if}
										</span>
										<div class="flex-1">
											<p class="font-semibold text-lg mb-1">
												{t('game.facts.guess_number', [originalIndex + 1, getGuessDisplayText(entry)], currentLocale)}
											</p>
											{#if entry.fact}
												<p class="text-text-muted prose prose-invert max-w-none" innerHTML={formatFact(entry.fact)}></p>
											{/if}
										</div>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>
