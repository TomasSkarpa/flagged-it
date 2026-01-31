<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { 
		GameSetupScreen, 
		GameHeader, 
		GameOverScreen,
		FactsGuessHistory
	} from '$lib/components/game';
	import { startFactsGame, submitGuess, skipRound, nextRound, type GuessHistoryEntry } from '$lib/api/factsGame';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import { getCountryNameForLocale } from '$lib/utils/countryNames';
	import { calculateCurrentRound } from '$lib/utils/gameUtils';
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
	let guessInputElement: HTMLInputElement | null = null;

	// Reactive translations
	$: currentLocale = $locale;
	$: factsTitle = t('game.facts.title', undefined, currentLocale);
	$: factsDescription = t('game.facts.description', undefined, currentLocale) || t('home.game.facts.description', undefined, currentLocale) || t('game.facts.guess_country', undefined, currentLocale);
	$: makeGuessText = t('game.facts.guess_country', undefined, currentLocale);
	$: enterCountryText = t('game.guessing.enter_country', undefined, currentLocale);
	$: guessText = t('game.guessing.guess', undefined, currentLocale);
	$: previousGuessesText = t('game.facts.previous_guesses', undefined, currentLocale);
	$: triesLeftText = t('game.facts.tries_left', undefined, currentLocale);
	$: tryLeftText = t('game.facts.try_left', undefined, currentLocale);
	$: triesLeftShortText = t('game.facts.tries_left_short', undefined, currentLocale);
	$: factHeaderText = t('game.facts.fact_header', undefined, currentLocale);
	$: guessCountryHeaderText = t('game.facts.guess_country_header', undefined, currentLocale);
	$: skipButtonText = t('game.facts.skip_button', undefined, currentLocale);
	$: skippedText = t('game.facts.skipped', undefined, currentLocale);
	$: skippedAnswerText = t('game.facts.skipped_answer', undefined, currentLocale);
	$: correctText = t('game.facts.correct', undefined, currentLocale);
	$: gameOverText = t('game.facts.game_over', undefined, currentLocale);
	$: wrongNextText = t('game.facts.wrong_next', undefined, currentLocale);
	$: wrongNoMoreText = t('game.facts.wrong_no_more', undefined, currentLocale);
	$: excellentMessage = t('game.over.excellent', undefined, currentLocale);
	$: playAgainText = t('game.over.play_again', undefined, currentLocale);

	// Re-fetch current fact when locale changes (to update country names)
	let previousLocale: string = currentLocale;
	$: if (gameStarted && sessionId && currentLocale !== previousLocale && !showFeedback && !isLoading && !gameFinished) {
		previousLocale = currentLocale;
		// Re-fetch the current fact with new locale to get translated country names
		nextRound(sessionId).then(result => {
			currentFact = result.currentFact || '';
			factNumber = result.factNumber || factNumber;
			triesLeft = result.triesLeft || triesLeft;
			score = result.score || score;
			total = result.total || total;
			guessInput = '';
			correctCountryName = '';
		}).catch(err => {
			console.error('Failed to reload fact with new locale:', err);
		});
	}
	
	// Reactive formatted fact - ensure it updates when currentFact changes
	$: formattedFact = currentFact && currentFact.trim() !== '' ? formatFact(currentFact) : '';
	$: hasFact = currentFact && currentFact.trim() !== '' && formattedFact && formattedFact.trim() !== '';

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
		currentFact = ''; // Reset current fact
		
		try {
			const result = await startFactsGame();
			
			// Set gameStarted first to ensure proper state
			gameStarted = true;
			gameFinished = false;
			
			sessionId = result.sessionId;
			
			// Ensure currentFact is properly set
			currentFact = result.currentFact || '';
			factNumber = result.factNumber || 1;
			triesLeft = result.triesLeft || 3;
			score = result.score || 0;
			total = result.total || 0;
			
			if (!currentFact || currentFact.trim() === '') {
				error = 'No fact was returned from the server. Please try again.';
				console.error('No currentFact in response. Result:', result);
				gameStarted = false;
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to start game';
			console.error('Start game error:', err);
			gameStarted = false;
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
			
			// Check if guess is valid (country exists)
			if (result.isValidGuess === false) {
				error = result.error || 'Country not found';
				isLoading = false;
				// Focus input so user can correct their guess
				setTimeout(() => {
					guessInputElement?.focus();
				}, 0);
				return;
			}
			
			isCorrect = result.isCorrect;
			triesLeft = result.triesLeft;
			score = result.score;
			total = result.total;
			guessHistory = result.guessHistory;

			if (result.isCorrect) {
				// Correct guess - show success message
				if (result.correctCountry) {
					const country = allCountries.find(c => c.cca2 === result.correctCountry!.cca2);
					correctCountryName = country ? getCountryNameForLocale(country) : result.correctCountry.name;
					correctCountryCca2 = result.correctCountry.cca2;
					statusMessage = correctText.replace('%s', correctCountryName);
				}
				showFeedback = true;
				isCorrect = true;

				// Wait 2 seconds, then start next round or finish game
				// Don't change fact here - let handleNextRound load the new fact
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
					correctCountryName = country ? getCountryNameForLocale(country) : result.correctCountry.name;
					correctCountryCca2 = result.correctCountry.cca2;
					statusMessage = gameOverText.replace('%s', correctCountryName).replace(' %s', '');
				}
				showFeedback = true;
				isCorrect = false;

				// Wait 2 seconds, then start next round or finish game
				// Don't change fact here - let handleNextRound load the new fact
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
				if (result.nextFact && result.nextFact.trim() !== '') {
					currentFact = result.nextFact;
					factNumber = result.factNumber !== undefined ? result.factNumber : factNumber + 1;
					statusMessage = wrongNextText;
				} else {
					statusMessage = wrongNoMoreText;
					console.warn('No nextFact in response, keeping current fact:', currentFact);
				}
				showFeedback = true;
				isCorrect = false;
				guessInput = '';
				
				// Clear feedback after 1.5 seconds to allow next guess
				setTimeout(() => {
					if (triesLeft > 0) {
						showFeedback = false;
						statusMessage = '';
					}
				}, 1500);
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

	async function handleSkip() {
		if (!sessionId || isLoading || showFeedback || gameFinished) return;

		isLoading = true;
		error = null;
		showFeedback = false;
		guessInput = ''; // Clear input

		try {
			const result = await skipRound(sessionId);
			
			isCorrect = false;
			triesLeft = result.triesLeft || 0;
			score = result.score || 0;
			total = result.total || 0;
			// Clear guess history when skipping - don't show Previous Guesses for skipped facts
			guessHistory = [];

			// Show skipped message with correct answer
			if (result.correctCountry) {
				const country = allCountries.find(c => c.cca2 === result.correctCountry!.cca2);
				correctCountryName = country ? getCountryNameForLocale(country) : result.correctCountry.name;
				correctCountryCca2 = result.correctCountry.cca2;
				statusMessage = skippedAnswerText.replace('%s', correctCountryName);
			} else {
				statusMessage = skippedText;
				correctCountryName = '';
				correctCountryCca2 = '';
			}
			showFeedback = true;

			// Wait 2 seconds, then start next round or finish game
			// Don't change fact here - let handleNextRound load the new fact
			setTimeout(async () => {
				showFeedback = false;
				statusMessage = '';
				if (result.isComplete) {
					gameFinished = true;
					gameStarted = false;
					isLoading = false;
				} else {
					await handleNextRound();
				}
			}, 2000);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to skip round';
			console.error('Skip error:', err);
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
		if (!fact || typeof fact !== 'string' || fact.trim() === '') {
			return '';
		}
		// Remove "Fact X: " prefix if present (backend includes it)
		const cleanedFact = fact.replace(/^Fact \d+:\s*/i, '');
		// Convert markdown-style bold **text** to HTML <strong>text</strong> with styling
		// Use fact-bold class that adapts to light/dark mode
		const formatted = cleanedFact.replace(/\*\*(.*?)\*\*/g, '<strong class="fact-bold font-bold">$1</strong>');
		return formatted || cleanedFact || fact;
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
		{:else if gameStarted && !gameFinished}
			<div class="space-y-6">
				<GameHeader
					{score}
					{total}
					currentRound={calculateCurrentRound(total, totalRounds)}
					{totalRounds}
				/>

				<!-- Fact Display and Input - Combined in one card-game -->
				<div class="card-game relative overflow-hidden">
					<!-- Header -->
					<div class="flex items-center justify-between mb-4">
						<div class="flex items-center gap-3">
							<span class="text-3xl">📚</span>
							<div>
								<h3 class="text-lg font-semibold text-sandy-light">{factHeaderText.replace('%d', (factNumber || 1).toString())}</h3>
								<p class="text-xs text-text-muted uppercase tracking-wide">{guessCountryHeaderText}</p>
							</div>
						</div>
						{#if triesLeft !== undefined && triesLeft > 0 && !showFeedback}
							<div class="px-3 py-1 bg-primary/20 rounded-full border border-primary/30">
								<span class="text-sm font-semibold text-primary">{triesLeft} {triesLeft === 1 ? tryLeftText : triesLeftShortText}</span>
							</div>
						{/if}
					</div>
					
					<!-- Fact Content -->
					<div class="p-6 bg-white/5 rounded-lg border border-white/10 min-h-[120px] flex items-center justify-center mb-6 {showFeedback ? 'opacity-0' : ''}">
						{#if hasFact}
							<div class="w-full">
								<div class="text-xl md:text-2xl text-sandy-light font-medium leading-relaxed text-center whitespace-normal break-words">
									{@html formattedFact}
								</div>
							</div>
						{:else}
							<div class="text-center py-4 w-full">
								<p class="text-text-muted mb-2">{currentFact ? 'Formatting fact...' : 'Loading fact...'}</p>
								{#if error}
									<p class="text-error text-sm mt-2">{error}</p>
								{/if}
							</div>
						{/if}
					</div>

					<!-- Guess Input Section -->
					{#if !gameFinished}
						<div class="w-full">
							<p class="text-center text-lg font-medium text-sandy-light mb-4">{makeGuessText}</p>
							{#if triesLeft > 0 && !showFeedback}
								<p class="text-center text-sm text-text-muted mb-3">{triesLeftText.replace('%d', triesLeft.toString())}</p>
							{/if}
							<div class="flex flex-col sm:flex-row gap-3">
								<input
									type="text"
									bind:this={guessInputElement}
									bind:value={guessInput}
									on:keypress={handleKeyPress}
									placeholder={enterCountryText}
									disabled={isLoading || showFeedback || gameFinished}
									class="flex-1 px-4 py-3 rounded-lg border-2 border-white/20 bg-white/5 text-sandy-light placeholder:text-text-muted focus:outline-none focus:border-primary transition-all disabled:opacity-50"
									autocomplete="off"
								/>
								<button
									on:click={handleSubmitGuess}
									disabled={isLoading || !guessInput.trim() || showFeedback || gameFinished}
									class="btn-primary px-8 py-3 font-semibold disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap"
								>
									{guessText}
								</button>
							</div>
							<div class="flex justify-center mt-4 pt-4 border-t border-white/10">
								<button
									on:click={handleSkip}
									disabled={isLoading || gameFinished}
									class="flex items-center gap-2 px-4 py-2 text-sm text-text-muted hover:text-sandy-light border border-white/20 hover:border-white/40 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed transition-all"
								>
									<span>⏭️</span>
									<span>{skipButtonText}</span>
								</button>
							</div>
							{#if error && !showFeedback}
								<div class="mt-3 p-3 bg-error/20 border border-error rounded-lg">
									<p class="text-error text-sm text-center font-medium">{error}</p>
								</div>
							{/if}
						</div>
					{/if}

					<!-- Feedback Overlay - Covers entire card-game -->
					{#if showFeedback && statusMessage}
						{@const isSkip = statusMessage.toLowerCase().includes('skipped')}
						<div class="feedback-overlay absolute inset-0 flex items-center justify-center rounded-card z-30 animate-fade-in {isCorrect === true ? 'overlay-success' : isCorrect === false && !isSkip ? 'overlay-error' : 'overlay-skip'}">
							<div class="text-center px-4">
								<div class="text-6xl mb-4 overlay-icon">
									{#if isCorrect === true}
										✓
									{:else if isSkip}
										⏭️
									{:else if isCorrect === false}
										✗
									{/if}
								</div>
								<p class="text-3xl font-bold overlay-text status-message mb-2">{statusMessage}</p>
								{#if correctCountryName && (isCorrect || triesLeft === 0 || isSkip)}
									<div class="flex items-center justify-center gap-2 mt-3">
										{#if correctCountryCca2}
											<img 
												src="/assets/twemoji_flags_cca2/{correctCountryCca2}.svg" 
												alt="{correctCountryName} flag"
												class="w-6 h-4 object-cover rounded"
											/>
										{/if}
										<span class="text-xl font-medium overlay-text correct-country-name">{correctCountryName}</span>
									</div>
								{/if}
							</div>
						</div>
					{/if}
				</div>

				<!-- Previous Guesses History -->
				{#if guessHistory.length > 0}
					{@const nonSkipGuesses = guessHistory.filter(entry => entry.guess.toLowerCase().trim() !== 'skip')}
					{#if nonSkipGuesses.length > 0}
						<FactsGuessHistory guesses={nonSkipGuesses} />
					{/if}
				{/if}
			</div>
		{/if}
	</div>
</div>

<style>
	/* Bold text in facts - visible in both light and dark mode */
	:global(.fact-bold) {
		color: var(--color-text-light);
	}

	:global(:root.light .fact-bold) {
		color: var(--color-text) !important;
	}

	/* Feedback Overlay Styles */
	/* Dark mode: dark overlay with light text */
	.feedback-overlay {
		background-color: rgba(0, 0, 0, 0.9);
	}

	.feedback-overlay.overlay-success {
		background-color: rgba(34, 197, 94, 0.9); /* success green */
	}

	.feedback-overlay.overlay-error {
		background-color: rgba(239, 68, 68, 0.9); /* error red */
	}

	.feedback-overlay.overlay-skip {
		background-color: rgba(0, 0, 0, 0.9) !important; /* dark overlay for skip in dark mode */
	}

	.overlay-text {
		color: white;
	}

	.overlay-icon {
		color: white;
	}

	/* Light mode: white-ish overlay with black text */
	:global(:root.light) .feedback-overlay {
		background-color: rgba(255, 255, 255, 0.95) !important;
	}

	:global(:root.light) .feedback-overlay.overlay-success {
		background-color: rgba(34, 197, 94, 0.9) !important;
	}

	:global(:root.light) .feedback-overlay.overlay-error {
		background-color: rgba(239, 68, 68, 0.9) !important;
	}

	:global(:root.light) .feedback-overlay.overlay-skip {
		background-color: rgba(255, 255, 255, 0.98) !important;
	}

	:global(:root.light) .overlay-text {
		color: #0F172A !important; /* black text */
	}

	:global(:root.light) .overlay-icon {
		color: #0F172A !important; /* black icon */
	}

	/* Light mode: correct country name should be black */
	:global(:root.light) .correct-country-name {
		color: #0F172A !important;
	}

	/* Light mode: status message should be black (but preserve success/error colors) */
	:global(:root.light) .status-message:not(.text-success):not(.text-error) {
		color: #0F172A !important;
	}
</style>
