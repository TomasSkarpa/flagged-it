<script lang="ts">
	import { onMount } from 'svelte';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import { 
		GameSetupScreen, 
		GameHeader, 
		GameContainer, 
		GameOverScreen 
	} from '$lib/components/game';
	import Keyboard from '$lib/components/ui/Keyboard.svelte';
	import type { KeyState } from '$lib/components/ui/keyboardTypes';
	import { HangmanGame } from '$lib/api/hangmanGame';
	import { getAllCountries } from '$lib/api/debug';
	import type { Country } from '$lib/types';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import { getKeyboardLayoutForLocale } from '$lib/utils/keyboardLayout';
	import { calculateCurrentRound } from '$lib/utils/gameUtils';
	import { triggerConfetti } from '$lib/utils/confetti';

	let countries: Country[] = [];
	let game: HangmanGame | null = null;
	let isLoading = false;
	let error: string | null = null;
	let gameStarted = false;
	let gameFinished = false;
	let selectedRegion = '';
	
	// Game state
	let currentWord: string = '';
	let guessedWord: string[] = [];
	let guessedLetters: string[] = [];
	let wrongGuesses = 0;
	let maxWrongGuesses = 6;
	let score = 0;
	let total = 0;
	let keyStates: Record<string, KeyState> = {};
	let currentCountry: Country | null = null;
	let showFeedback = false;
	let isCorrect = false;
	let correctAnswer = '';
	let customMessage = '';

	// Reactive translations
	$: currentLocale = $locale;
	$: hangmanTitle = t('game.hangman.title', undefined, currentLocale);
	$: hangmanDescription = t('game.hangman.title', undefined, currentLocale);
	$: guessCountryText = t('game.hangman.guess_country', undefined, currentLocale);
	$: alreadyGuessedText = t('game.hangman.already_guessed', undefined, currentLocale);
	$: congratulationsText = t('game.hangman.congratulations', undefined, currentLocale);
	
	// Functions for translations with parameters
	function getGameOverText(word: string): string {
		return t('game.hangman.game_over', [word], currentLocale);
	}
	function getWrongGuessesText(wrong: number, max: number): string {
		return t('game.hangman.wrong_guesses', [wrong, max], currentLocale);
	}
	function getLettersWordsText(letters: number, words: number, wordForm: string): string {
		return t('game.hangman.letters_words', [letters, words, wordForm], currentLocale);
	}
	
	// Keyboard layout based on locale
	$: keyboardLayout = getKeyboardLayoutForLocale(currentLocale);

	onMount(async () => {
		isLoading = true;
		try {
			const result = await getAllCountries();
			countries = result.countries;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load countries';
			console.error('Error loading countries:', err);
		} finally {
			isLoading = false;
		}
	});

	function handleStartGame(event: CustomEvent<{ region?: string; category?: string; [key: string]: any }>) {
		selectedRegion = event.detail.region ?? '';
		
		// Filter countries by region if selected
		let filteredCountries = countries;
		if (selectedRegion) {
			filteredCountries = countries.filter(c => c.region === selectedRegion);
		}

		if (filteredCountries.length === 0) {
			error = 'No countries available for selected region';
			return;
		}

		// Create new game instance with current locale
		game = new HangmanGame(filteredCountries, currentLocale);
		game.newRound();
		
		// Update state
		const state = game.getState();
		currentWord = state.currentWord;
		guessedWord = state.guessedWord;
		guessedLetters = state.guessedLetters;
		wrongGuesses = state.wrongGuesses;
		score = state.score;
		total = state.total;
		currentCountry = state.country;
		keyStates = {};
		showFeedback = false;
		isCorrect = false;
		correctAnswer = '';
		customMessage = '';
		gameStarted = true;
		gameFinished = false;
	}

	// Track previous locale to detect changes
	let previousLocale = currentLocale;
	
	// Restart entire game when locale changes
	$: if (game && currentLocale && currentLocale !== previousLocale && gameStarted && !gameFinished) {
		previousLocale = currentLocale;
		
		// Filter countries by region if selected
		let filteredCountries = countries;
		if (selectedRegion) {
			filteredCountries = countries.filter(c => c.region === selectedRegion);
		}

		if (filteredCountries.length > 0) {
			// Create new game instance with current locale (full restart)
			game = new HangmanGame(filteredCountries, currentLocale);
			game.newRound();
			
			// Reset all game state
			const state = game.getState();
			currentWord = state.currentWord;
			guessedWord = state.guessedWord;
			guessedLetters = state.guessedLetters;
			wrongGuesses = state.wrongGuesses;
			score = 0; // Reset score
			total = 0; // Reset total
			currentCountry = state.country;
			keyStates = {};
			showFeedback = false;
			isCorrect = false;
			correctAnswer = '';
			customMessage = '';
			gameStarted = true;
			gameFinished = false;
		} else {
			error = 'No countries available for selected region';
		}
	}

	function handleKeyPress(event: CustomEvent<{ key: string }>) {
		if (!game || gameFinished || showFeedback) return;

		const letter = event.detail.key;
		const result = game.makeGuess(letter);

		if (!result.isValidGuess) {
			if (result.error === 'Letter already guessed') {
				// Show brief feedback for already guessed (optional - could be a toast)
				return;
			}
			return;
		}

		// Update key states
		if (result.isInWord) {
			keyStates[letter] = 'correct';
		} else {
			keyStates[letter] = 'incorrect';
		}
		keyStates = { ...keyStates }; // Trigger reactivity

		// Update game state
		guessedWord = result.guessedWord;
		wrongGuesses = result.wrongGuesses;
		score = result.score;
		total = result.total;

		// Check game end
		if (result.isWon) {
			isCorrect = true;
			correctAnswer = '';
			customMessage = '';
			showFeedback = true;
			
			// Trigger confetti for win
			triggerConfetti({
				particleCount: 50,
				spread: 60,
				origin: { x: 0.5, y: 0.4 },
				duration: 2500
			});
			
			setTimeout(() => {
				showFeedback = false;
				if (result.isComplete) {
					gameFinished = true;
					gameStarted = false;
				} else {
					startNewRound();
				}
			}, 2000);
		} else if (result.isGameOver) {
			isCorrect = false;
			// Format the message: "Game Over! The word was: X"
			const revealedWord = result.revealedWord || currentWord;
			customMessage = getGameOverText(revealedWord);
			correctAnswer = '';
			showFeedback = true;
			setTimeout(() => {
				showFeedback = false;
				if (result.isComplete) {
					gameFinished = true;
					gameStarted = false;
				} else {
					startNewRound();
				}
			}, 2000);
		}
	}

	function startNewRound() {
		if (!game) return;
		
		game.newRound();
		const state = game.getState();
		currentWord = state.currentWord;
		guessedWord = state.guessedWord;
		guessedLetters = state.guessedLetters;
		wrongGuesses = state.wrongGuesses;
		currentCountry = state.country;
		keyStates = {};
		showFeedback = false;
		isCorrect = false;
		correctAnswer = '';
		customMessage = '';
	}

	function handlePlayAgain() {
		gameStarted = false;
		gameFinished = false;
		game = null;
		score = 0;
		total = 0;
		keyStates = {};
		showFeedback = false;
		isCorrect = false;
		correctAnswer = '';
		customMessage = '';
	}

	// Calculate hangman image opacity/stage based on wrong guesses
	$: hangmanStage = Math.min(wrongGuesses / maxWrongGuesses, 1);
	$: displayWord = guessedWord.join(' ');
</script>

<svelte:head>
	<title>{hangmanTitle} - Flagged It</title>
	<meta name="description" content={hangmanDescription} />
</svelte:head>

{#if isLoading}
	<div class="min-h-screen flex items-center justify-center">
		<LoadingSpinner />
	</div>
{:else if error}
	<div class="min-h-screen flex items-center justify-center">
		<div class="card-game text-center">
			<p class="text-error text-xl mb-4">{error}</p>
			<button class="btn-primary" on:click={() => window.location.reload()}>
				{t('common.retry', undefined, currentLocale) || 'Retry'}
			</button>
		</div>
	</div>
{:else if !gameStarted && !gameFinished}
	<GameSetupScreen
		title={hangmanTitle}
		emoji="🎯"
		description={hangmanDescription}
		{isLoading}
		{error}
		showRegionSelector={true}
		on:start={handleStartGame}
	/>
{:else if gameFinished}
	<GameOverScreen
		title={t('game.over.title', undefined, currentLocale)}
		score={score}
		totalRounds={5}
		excellentMessage={t('game.over.excellent', undefined, currentLocale)}
		goodMessage={t('game.over.good', undefined, currentLocale)}
		encourageMessage={t('game.over.encourage', undefined, currentLocale)}
		on:playAgain={handlePlayAgain}
	/>
{:else}
	<div class="min-h-screen py-2 px-4">
		<div class="max-w-4xl mx-auto">
			<GameHeader
				score={score}
				total={5}
				currentRound={calculateCurrentRound(total, 5)}
				totalRounds={5}
			/>

			<GameContainer
				question={guessCountryText}
				{showFeedback}
				{isCorrect}
				{correctAnswer}
				{customMessage}
			>
				<div slot="content" class="flex flex-col items-center gap-3 mb-2">
					<!-- Hangman Image -->
					<div class="hangman-container relative w-48 h-56 md:w-56 md:h-64 flex items-center justify-center">
						<img 
							src="/assets/iconography/hangman.svg" 
							alt="Hangman"
							class="hangman-image w-full h-full object-contain"
							style="opacity: {Math.max(0.1, hangmanStage)}; filter: {wrongGuesses >= maxWrongGuesses ? 'grayscale(100%)' : 'none'};"
						/>
						{#if wrongGuesses >= maxWrongGuesses}
							<div class="absolute inset-0 flex items-center justify-center">
								<span class="text-4xl md:text-5xl">💀</span>
							</div>
						{/if}
					</div>

					<!-- Word Display -->
					<div class="word-display text-center">
						<div class="text-3xl md:text-4xl font-mono font-bold text-sandy-light mb-2 tracking-wider">
							{displayWord}
						</div>
						{#if currentCountry}
							<p class="text-text-muted text-xs">
								{getLettersWordsText(
									currentWord.replace(/\s/g, '').length,
									currentWord.split(' ').length,
									currentWord.split(' ').length === 1 
										? t('game.hangman.word', undefined, currentLocale)
										: t('game.hangman.words', undefined, currentLocale)
								)}
							</p>
						{/if}
					</div>

					<!-- Wrong Guesses Counter -->
					<div class="wrong-guesses text-center">
						<p class="text-text-muted text-xs">
							{getWrongGuessesText(wrongGuesses, maxWrongGuesses)}
						</p>
					</div>
				</div>

				<!-- Keyboard -->
				<div slot="answers" class="keyboard-container mt-2">
					<Keyboard
						layout={keyboardLayout}
						{keyStates}
						disabled={gameFinished || wrongGuesses >= maxWrongGuesses || showFeedback}
						on:keypress={handleKeyPress}
					/>
				</div>
			</GameContainer>
		</div>
	</div>
{/if}

<style>
	.hangman-container {
		min-height: 200px;
	}

	.hangman-image {
		filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.3));
		transition: opacity 0.3s ease;
	}

	.word-display {
		min-height: 60px;
	}
</style>
