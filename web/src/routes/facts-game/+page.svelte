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
	
	// Reactive formatted fact - ensure it updates when currentFact changes
	$: formattedFact = currentFact && currentFact.trim() !== '' ? formatFact(currentFact) : '';
	$: hasFact = currentFact && currentFact.trim() !== '' && formattedFact && formattedFact.trim() !== '';
	
	// Debug: Log when currentFact changes
	$: if (browser) {
		console.log('currentFact reactive:', currentFact?.substring(0, 50), 'formatted:', formattedFact?.substring(0, 50), 'hasFact:', hasFact);
	}

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
			console.log('Starting facts game...');
			const result = await startFactsGame();
			console.log('Facts game started:', result);
			
			// Set gameStarted first to ensure proper state
			gameStarted = true;
			gameFinished = false;
			
			sessionId = result.sessionId;
			
			// Ensure currentFact is properly set - check multiple possible fields
			const factValue = result.currentFact || result.fact || '';
			currentFact = factValue;
			factNumber = result.factNumber || 1;
			triesLeft = result.triesLeft || 3;
			score = result.score || 0;
			total = result.total || 0;
			
			console.log('currentFact set to:', currentFact);
			console.log('factNumber set to:', factNumber);
			console.log('gameStarted:', gameStarted);
			console.log('Will display fact?', currentFact && currentFact.trim() !== '');
			
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
					console.log('Updated currentFact with nextFact:', currentFact);
					console.log('Updated factNumber:', factNumber);
				} else if (result.currentFact && result.currentFact.trim() !== '') {
					// Fallback: use currentFact if nextFact is not provided
					currentFact = result.currentFact;
					factNumber = result.factNumber !== undefined ? result.factNumber : factNumber + 1;
					statusMessage = wrongNextText;
					console.log('Updated currentFact from currentFact field:', currentFact);
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

		try {
			console.log('Skipping round...');
			const result = await skipRound(sessionId);
			console.log('Skip result:', result);
			
			isCorrect = false;
			triesLeft = result.triesLeft;
			score = result.score;
			total = result.total;
			guessHistory = result.guessHistory;

			// Show skipped message with correct answer
			if (result.correctCountry) {
				const country = allCountries.find(c => c.cca2 === result.correctCountry!.cca2);
				correctCountryName = country ? getCountryName(country, currentLocale) : result.correctCountry.name;
				correctCountryCca2 = result.correctCountry.cca2;
				statusMessage = `Skipped! The answer was ${correctCountryName}`;
			} else {
				statusMessage = 'Skipped!';
			}
			showFeedback = true;

			// Wait 2 seconds, then start next round or finish game
			setTimeout(async () => {
				if (result.isComplete) {
					gameFinished = true;
					gameStarted = false;
				} else {
					await handleNextRound();
				}
			}, 2000);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to skip round';
			console.error('Skip error:', err);
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
					currentRound={total + 1}
					{totalRounds}
				/>

				<!-- Current Fact Display - Always show when game is started -->
				<div class="card-game relative">
					<div class="flex items-center justify-between mb-4">
						<div class="flex items-center gap-3">
							<span class="text-3xl">📚</span>
							<div>
								<h3 class="text-lg font-semibold text-sandy-light">Fact #{factNumber || 1}</h3>
								<p class="text-xs text-text-muted uppercase tracking-wide">Guess the Country</p>
							</div>
						</div>
						{#if triesLeft !== undefined && triesLeft > 0 && !showFeedback}
							<div class="px-3 py-1 bg-primary/20 rounded-full border border-primary/30">
								<span class="text-sm font-semibold text-primary">{triesLeft} {triesLeft === 1 ? 'try' : 'tries'} left</span>
							</div>
						{/if}
					</div>
					
					<div class="p-6 bg-white/5 rounded-lg border border-white/10 min-h-[120px] flex items-center justify-center">
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
								{#if browser && currentFact}
									<p class="text-xs text-text-muted mt-2">Debug: fact length={currentFact.length}, formatted length={formattedFact?.length || 0}</p>
								{/if}
							</div>
						{/if}
					</div>
				</div>

				<!-- Feedback Message -->
				{#if showFeedback && statusMessage}
					{@const isSkip = statusMessage.toLowerCase().includes('skipped')}
					<div class={`card-game ${isCorrect === true ? 'bg-success/20 border-success' : isCorrect === false && !isSkip ? 'bg-error/20 border-error' : isSkip ? 'bg-white/10 border-white/30' : ''}`}>
						<div class="flex items-center justify-center gap-3 mb-2">
							{#if isCorrect === true}
								<span class="text-3xl">✓</span>
							{:else if isSkip}
								<span class="text-3xl">⏭️</span>
							{:else if isCorrect === false}
								<span class="text-3xl">✗</span>
							{/if}
							<p class="text-center font-semibold text-lg {isCorrect === true ? 'text-success' : isCorrect === false && !isSkip ? 'text-error' : isSkip ? 'text-text-light' : 'text-sandy-light'}">
								{statusMessage}
							</p>
						</div>
						{#if correctCountryName && (isCorrect || triesLeft === 0 || isSkip)}
							<div class="flex items-center justify-center gap-2 mt-3 pt-3 border-t border-white/10">
								{#if correctCountryCca2}
									<img 
										src="/assets/twemoji_flags_cca2/{correctCountryCca2}.svg" 
										alt="{correctCountryName} flag"
										class="w-6 h-4 object-cover rounded"
									/>
								{/if}
								<span class="text-text-light font-medium">{correctCountryName}</span>
							</div>
						{/if}
					</div>
				{/if}

				<!-- Guess Input Section -->
				{#if triesLeft > 0 && (!isCorrect || !showFeedback)}
					<div class="card-game">
						<p class="text-center text-lg font-medium text-sandy-light mb-4">{makeGuessText}</p>
						{#if triesLeft > 0 && !showFeedback}
							<p class="text-center text-sm text-text-muted mb-3">{triesLeftText.replace('%d', triesLeft.toString())}</p>
						{/if}
						<div class="flex flex-col sm:flex-row gap-3">
							<input
								type="text"
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
								disabled={isLoading || showFeedback || gameFinished}
								class="flex items-center gap-2 px-4 py-2 text-sm text-text-muted hover:text-sandy-light border border-white/20 hover:border-white/40 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed transition-all"
							>
								<span>⏭️</span>
								<span>I don't know / Skip</span>
							</button>
						</div>
						{#if error && !showFeedback}
							<p class="text-error text-sm text-center mt-3">{error}</p>
						{/if}
					</div>
				{/if}

				<!-- Previous Guesses History -->
				{#if guessHistory.length > 0}
					<FactsGuessHistory guesses={guessHistory} />
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

	:global(:root.light) .fact-bold {
		color: var(--color-text) !important;
	}
</style>
