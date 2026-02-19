<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import { 
		GameSetupScreen, 
		GameHeader, 
		AnswerGrid,
		GameOverScreen,
		DifficultySelector
	} from '$lib/components/game';
	import type { DifficultyOption } from '$lib/components/game/DifficultySelector.svelte';
	import { startCapitalGame, getCapitalQuestion, submitCapitalAnswer } from '$lib/api/capitalGame';
	import type { CapitalQuestion } from '$lib/api/capitalGame';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import { getAllCountries } from '$lib/api/data';
	import { getCountryNameForLocale } from '$lib/utils/countryNames';
	import { calculateCurrentRound } from '$lib/utils/gameUtils';
	import type { Country } from '$lib/types';

	type DifficultyMode = 'regular' | 'intermediate';

	let sessionId: string | null = null;
	let currentQuestion: CapitalQuestion | null = null;
	let selectedAnswer: string | null = null;
	let correctCapital: string = '';
	let score = 0;
	let total = 0;
	let isLoading = false;
	let error: string | null = null;
	let showFeedback = false;
	let isCorrect = false;
	let selectedRegion = '';
	let gameStarted = false;
	let gameFinished = false;
	const totalRounds = 10;
	
	// Difficulty mode state
	let difficultyMode: DifficultyMode = 'regular';

	// Text input for intermediate mode
	let guessInput = '';
	let guessInputElement: HTMLInputElement | null = null;
	
	// Store all countries for translation lookup and capital matching
	let allCountries: Country[] = [];
	let countriesLoaded = false;

	// Reactive difficulty modes with translations
	let difficultyModes: DifficultyOption[] = [];
	$: difficultyModes = [
		{ 
			value: 'regular', 
			label: t('game.capital.difficulty.regular', undefined, currentLocale), 
			description: t('game.capital.difficulty.regular.desc', undefined, currentLocale) 
		},
		{ 
			value: 'intermediate', 
			label: t('game.capital.difficulty.intermediate', undefined, currentLocale), 
			description: t('game.capital.difficulty.intermediate.desc', undefined, currentLocale) 
		}
	];

	// Helper to check if mode uses text input
	function usesTextInput(mode: DifficultyMode): boolean {
		return mode === 'intermediate';
	}

	// Helper to find capital name by matching input
	function findCapitalByName(input: string, countryCca2: string): string | null {
		if (!input || !allCountries.length) return null;
		
		const inputLower = input.trim().toLowerCase();
		if (inputLower === '') return null;

		// Find the country for this question
		const country = allCountries.find(c => c.cca2 === countryCca2);
		if (!country || !country.capital || country.capital.length === 0) return null;

		// Check if input matches any capital name (case-insensitive)
		for (const capital of country.capital) {
			if (capital.toLowerCase() === inputLower) {
				return capital; // Return the original capital name
			}
		}

		return null;
	}

	// Load countries on mount for translation and capital matching
	onMount(async () => {
		if (!browser || countriesLoaded) return;
		try {
			const result = await getAllCountries();
			allCountries = result.countries;
			countriesLoaded = true;
		} catch (err) {
			console.error('Failed to load countries for translation:', err);
			// Continue without translations - will use English names
		}
	});

	// Reactive translations
	$: currentLocale = $locale;
	$: gameTitle = t('game.capital.title', undefined, currentLocale);
	$: gameDescription = t('game.capital.description', undefined, currentLocale);
	// Translate country name in question text
	$: translatedCountryName = currentQuestion && allCountries.length > 0
		? (() => {
			const country = allCountries.find(c => c.cca2 === currentQuestion!.countryCca2);
			return country ? getCountryNameForLocale(country) : currentQuestion!.countryName;
		})()
		: (currentQuestion?.countryName || '');
	$: questionText = t('game.capital.question', { Country: translatedCountryName }, currentLocale);
	$: excellentMessage = t('game.over.excellent', undefined, currentLocale);
	$: loadingCountriesText = t('library.loading', undefined, currentLocale);
	$: failedToStartGameError = t('common.error', undefined, currentLocale);

	async function handleStartGame(event: CustomEvent<{ region?: string }>) {
		isLoading = true;
		error = null;
		gameFinished = false;
		score = 0;
		total = 0;
		selectedRegion = event.detail.region || '';
		
		try {
			const result = await startCapitalGame(selectedRegion || '');
			sessionId = result.sessionId;
			currentQuestion = result.question;
			gameStarted = true;
			selectedAnswer = null;
			showFeedback = false;
		} catch (err) {
			error = err instanceof Error ? err.message : failedToStartGameError;
			console.error('Start game error:', err);
		} finally {
			isLoading = false;
		}
	}

	async function handleSelectAnswer(event: CustomEvent<{ answer?: string; country?: any }>) {
		if (showFeedback || !sessionId || !currentQuestion) return;
		
		const answer = event.detail.answer || event.detail.country?.name?.common;
		if (!answer) return;
		
		selectedAnswer = answer;
		isLoading = true;
		error = null;
		
		try {
			const result = await submitCapitalAnswer(sessionId, currentQuestion.questionId, answer);
			isCorrect = result.correct;
			correctCapital = result.correctCapital;
			score = result.score;
			total = result.total;
			showFeedback = true;
			
			setTimeout(async () => {
				showFeedback = false;
				if (total >= totalRounds || result.finished) {
					gameFinished = true;
					gameStarted = false;
				} else {
					await loadNextQuestion();
				}
			}, 1200);
		} catch (err) {
			error = err instanceof Error ? err.message : failedToStartGameError;
			console.error('Submit answer error:', err);
		} finally {
			isLoading = false;
		}
	}

	async function loadNextQuestion() {
		if (!sessionId) return;
		
		isLoading = true;
		error = null;
		selectedAnswer = null;
		guessInput = '';
		
		try {
			const question = await getCapitalQuestion(sessionId);
			currentQuestion = question;
			// Reset selectedAnswer when new question loads
			selectedAnswer = null;
			// Focus input if using text input mode
			if (usesTextInput(difficultyMode) && guessInputElement) {
				setTimeout(() => {
					guessInputElement?.focus();
				}, 0);
			}
		} catch (err) {
			error = err instanceof Error ? err.message : failedToStartGameError;
			console.error('Load question error:', err);
		} finally {
			isLoading = false;
		}
	}

	// Track previous question ID to reset selectedAnswer when question changes
	let previousQuestionId: string | null = null;
	$: if (currentQuestion?.questionId && currentQuestion.questionId !== previousQuestionId) {
		previousQuestionId = currentQuestion.questionId;
		// Reset selectedAnswer when question changes (but not during feedback)
		if (!showFeedback) {
			selectedAnswer = null;
		}
	}

	// Re-fetch current question when locale changes (to update country and capital names)
	let previousLocale: string = currentLocale;
	$: if (gameStarted && currentQuestion && currentLocale !== previousLocale && !showFeedback && !isLoading) {
		previousLocale = currentLocale;
		// Re-fetch the current question with new locale to get translated names
		if (sessionId && currentQuestion) {
			getCapitalQuestion(sessionId).then(question => {
				currentQuestion = question;
				selectedAnswer = null;
				// Clear correctCapital - it will be set again when user submits next answer
				correctCapital = '';
			}).catch(err => {
				console.error('Failed to reload question with new locale:', err);
			});
		}
	}

	async function handleSubmitTextGuess() {
		if (showFeedback || !sessionId || !currentQuestion || !guessInput.trim() || isLoading) return;

		const capitalName = guessInput.trim();
		isLoading = true;
		error = null;

		try {
			// Find capital by name using helper function
			const matchedCapital = findCapitalByName(capitalName, currentQuestion.countryCca2);

			if (!matchedCapital) {
				error = t('game.guessing.capital_not_found', undefined, currentLocale) || 'Capital not found!';
				isLoading = false;
				setTimeout(() => {
					guessInputElement?.focus();
				}, 0);
				return;
			}

			const result = await submitCapitalAnswer(sessionId, currentQuestion.questionId, matchedCapital);
			isCorrect = result.correct;
			correctCapital = result.correctCapital;
			score = result.score;
			total = result.total;
			showFeedback = true;
			guessInput = '';

			setTimeout(async () => {
				showFeedback = false;
				if (result.finished || total >= totalRounds) {
					gameFinished = true;
					gameStarted = false;
				} else {
					await loadNextQuestion();
				}
			}, 2000);
		} catch (err) {
			error = err instanceof Error ? err.message : failedToStartGameError;
			console.error('Submit answer error:', err);
		} finally {
			isLoading = false;
		}
	}

	function handleKeyPress(event: KeyboardEvent) {
		if (event.key === 'Enter' && !isLoading) {
			handleSubmitTextGuess();
		}
	}

	function handleDifficultySelect(event: CustomEvent<{ value: string }>) {
		difficultyMode = event.detail.value as DifficultyMode;
	}

	function handlePlayAgain() {
		gameStarted = false;
		gameFinished = false;
		currentQuestion = null;
		sessionId = null;
		score = 0;
		total = 0;
		selectedAnswer = null;
		showFeedback = false;
		guessInput = '';
		difficultyMode = 'regular'; // Reset to default
	}

</script>

<svelte:head>
	<title>{gameTitle} - Flagged It</title>
</svelte:head>

<div class="min-h-screen p-4 md:p-8">
	<div class="max-w-4xl mx-auto">
		{#if !gameStarted && !gameFinished}
			<GameSetupScreen
				title={gameTitle}
				emoji="🏛️"
				description={gameDescription}
				{isLoading}
				{error}
				bind:selectedRegion
				on:start={handleStartGame}
			>
				<div slot="options" class="mb-8">
					<div class="max-w-2xl mx-auto">
						<DifficultySelector
							options={difficultyModes}
							selected={difficultyMode}
							on:select={handleDifficultySelect}
						/>
					</div>
				</div>
			</GameSetupScreen>
		{:else if gameFinished}
			<GameOverScreen
				{score}
				{totalRounds}
				excellentMessage={excellentMessage}
				on:playAgain={handlePlayAgain}
			/>
		{:else if currentQuestion}
			<GameHeader
				{score}
				{total}
				currentRound={calculateCurrentRound(total, totalRounds)}
				{totalRounds}
			/>
			
			<div 
				class="card-game relative overflow-hidden flag-background-container"
				style="background-image: url('/assets/twemoji_flags_cca2/{currentQuestion.countryCca2}.svg');"
				data-flag-background="true"
			>
				<div class="flag-overlay"></div>
				
				<div class="relative z-10">
					<h2 class="text-2xl md:text-3xl font-bold text-white dark:text-white text-slate-900 text-center mb-8 drop-shadow-lg capital-question">
						{questionText}
					</h2>
					
					<div class="mb-10"></div>
					
					{#if usesTextInput(difficultyMode)}
						<!-- Text input for intermediate mode -->
						<div class="w-full">
							<div class="flex flex-col sm:flex-row gap-3">
								<input
									type="text"
									bind:this={guessInputElement}
									bind:value={guessInput}
									on:keypress={handleKeyPress}
									placeholder={t('game.guessing.enter_capital', undefined, currentLocale)}
									disabled={isLoading || showFeedback || gameFinished}
									class="flex-1 px-4 py-3 rounded-lg border-2 border-white/20 bg-white/5 text-sandy-light placeholder:text-text-muted focus:outline-none focus:border-primary transition-all disabled:opacity-50"
									autocomplete="off"
								/>
								<button
									on:click={handleSubmitTextGuess}
									disabled={isLoading || !guessInput.trim() || showFeedback || gameFinished}
									class="btn-primary px-8 py-3 font-semibold disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap"
								>
									{t('game.guessing.guess', undefined, currentLocale)}
								</button>
							</div>
							{#if error && !showFeedback}
								<div class="mt-3 p-3 bg-error/20 border border-error rounded-lg">
									<p class="text-error text-sm text-center font-medium">{error}</p>
								</div>
							{/if}
						</div>
					{:else}
						<!-- Multiple choice for regular mode -->
						<AnswerGrid
							options={currentQuestion.options}
							{selectedAnswer}
							correctAnswer={correctCapital}
							{showFeedback}
							{isCorrect}
							disabled={isLoading}
							on:select={handleSelectAnswer}
							columns={1}
						/>
					{/if}
					
					{#if error && !usesTextInput(difficultyMode)}
						<div class="mt-6 p-4 bg-error/20 border border-error rounded-lg">
							<p class="text-error font-semibold text-center">{error}</p>
						</div>
					{/if}
				</div>
				
				<!-- Feedback Overlay -->
				{#if showFeedback}
					<div 
						class="absolute inset-0 flex items-center justify-center rounded-card animate-fade-in z-20
							{isCorrect ? 'bg-success/50' : 'bg-error/50'}"
					>
						<div class="text-center">
							<div class="text-6xl mb-4">{isCorrect ? '✓' : '✗'}</div>
							<p class="text-3xl font-bold text-white">{isCorrect ? t('game.correct_short', undefined, currentLocale) : t('game.wrong_short', undefined, currentLocale)}</p>
							{#if !isCorrect && correctCapital}
								<p class="text-xl text-white/90 mt-2">{correctCapital}</p>
							{/if}
						</div>
					</div>
				{/if}
			</div>
		{:else if isLoading}
			<div class="flex flex-col items-center justify-center py-20">
				<LoadingSpinner />
				<p class="mt-4 text-text-muted">{t('common.loading', undefined, currentLocale)}</p>
			</div>
		{/if}
	</div>
</div>

<style>
	.flag-background-container {
		background-size: cover;
		background-position: center;
		background-repeat: no-repeat;
		min-height: 500px;
	}
	
	.flag-overlay {
		position: absolute;
		inset: 0;
		background: linear-gradient(
			to bottom,
			rgba(10, 14, 39, 0.85) 0%,
			rgba(10, 14, 39, 0.75) 50%,
			rgba(10, 14, 39, 0.9) 100%
		);
		border-radius: 1.5rem; /* Match rounded-card from Tailwind config */
		z-index: 1;
	}

	/* Light mode: white overlay to create light background for dark text */
	:global(:root.light) .flag-overlay {
		background: linear-gradient(
			to bottom,
			rgba(255, 255, 255, 0.7) 0%,
			rgba(255, 255, 255, 0.6) 50%,
			rgba(255, 255, 255, 0.75) 100%
		);
	}
	
	@media (min-width: 768px) {
		.flag-background-container {
			min-height: 600px;
		}
	}

	/* Light mode: dark text for question on flag background */
	:global(:root.light) .capital-question {
		color: #0F172A !important; /* Dark slate - very visible */
		text-shadow: 0 1px 2px rgba(255, 255, 255, 0.8);
	}

	/* Light mode: button styles for flag background */
	/* Use data attribute for more reliable targeting */
	:global(:root.light) [data-flag-background="true"] :global(button) {
		text-shadow: none !important;
		transition: all 0.2s ease !important;
	}

	/* Regular buttons - must override Tailwind classes */
	:global(:root.light) [data-flag-background="true"] :global(button:not(.bg-success):not(.bg-error):not(.bg-primary\/30)) {
		background-color: rgba(255, 255, 255, 0.95) !important;
		border-color: rgba(15, 23, 42, 0.3) !important;
		color: #0F172A !important;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1) !important;
	}

	/* Hover state - must override Tailwind hover:border-accent hover:bg-accent/10 */
	:global(:root.light) [data-flag-background="true"] :global(button:not(.bg-success):not(.bg-error):not(.bg-primary\/30)):hover:not(:disabled) {
		background-color: rgba(6, 182, 212, 0.25) !important;
		border-color: rgba(6, 182, 212, 0.7) !important;
		color: #0F172A !important;
		box-shadow: 0 4px 12px rgba(6, 182, 212, 0.3) !important;
		transform: translateY(-2px) !important;
	}

	/* Selected button (primary) */
	:global(:root.light) [data-flag-background="true"] :global(button.bg-primary\/30) {
		background-color: rgba(79, 70, 229, 0.15) !important;
		border-color: rgba(79, 70, 229, 0.4) !important;
		color: #0F172A !important;
	}

	/* Selected button hover */
	:global(:root.light) [data-flag-background="true"] :global(button.bg-primary\/30):hover:not(:disabled) {
		background-color: rgba(79, 70, 229, 0.35) !important;
		border-color: rgba(79, 70, 229, 0.8) !important;
		color: #0F172A !important;
		box-shadow: 0 4px 12px rgba(79, 70, 229, 0.4) !important;
		transform: translateY(-2px) !important;
	}
</style>
