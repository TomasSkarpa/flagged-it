<script lang="ts">
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import { 
		GameSetupScreen, 
		GameHeader, 
		GameContainer, 
		AnswerGrid, 
		GameOverScreen,
		DifficultySelector
	} from '$lib/components/game';
	import type { DifficultyOption } from '$lib/components/game/DifficultySelector.svelte';
	import { startGame, getQuestion, submitAnswer } from '$lib/api/flagGame';
	import type { Country } from '$lib/types';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import { getCountryNameForLocale, findCountryByName } from '$lib/utils/countryNames';
	import { calculateCurrentRound } from '$lib/utils/gameUtils';
	import { getAllCountries } from '$lib/api/data';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';

	type DifficultyMode = 'regular' | 'expert';

	let sessionId: string | null = null;
	let currentQuestion: any = null;
	let selectedAnswer: string | null = null;
	let correctAnswer: string = '';
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

	// Text input for expert mode
	let guessInput = '';
	let guessInputElement: HTMLInputElement | null = null;
	let allCountries: Country[] = [];
	let countriesLoaded = false;
	let flagImageLoaded = false;

	// Reactive translations - will update when locale changes
	$: currentLocale = $locale;
	$: flagGameTitle = t('game.flag.setup.title', undefined, currentLocale);
	$: flagGameDescription = t('game.flag.setup.description', undefined, currentLocale);
	$: flagSeoTitle = t('game.flag.title', undefined, currentLocale);
	$: flagPageDescription = t('home.game.flag.description', undefined, currentLocale);
	$: flagQuestionText = t('game.flag.question', undefined, currentLocale);
	$: excellentMessage = t('game.flag.over.excellent', undefined, currentLocale);

	// Reactive difficulty modes with translations
	let difficultyModes: DifficultyOption[] = [];
	$: difficultyModes = [
		{ 
			value: 'regular', 
			label: t('game.flag.difficulty.regular', undefined, currentLocale), 
			description: t('game.flag.difficulty.regular.desc', undefined, currentLocale) 
		},
		{ 
			value: 'expert', 
			label: t('game.flag.difficulty.expert', undefined, currentLocale), 
			description: t('game.flag.difficulty.expert.desc', undefined, currentLocale) 
		}
	];

	// Helper to check if mode uses text input
	function usesTextInput(mode: DifficultyMode): boolean {
		return mode === 'expert';
	}

	onMount(async () => {
		if (!browser) return;
		try {
			const result = await getAllCountries();
			allCountries = result.countries;
			countriesLoaded = true;
		} catch (err) {
			console.warn('Failed to load countries for translation:', err);
		}
	});

	async function handleStartGame(event: CustomEvent<{ region?: string; roundCount?: number }>) {
		isLoading = true;
		error = null;
		gameFinished = false;
		score = 0;
		total = 0;
		selectedRegion = event.detail.region || '';
		flagImageLoaded = false;
		const opts = event.detail.roundCount != null && event.detail.roundCount > 0
			? { roundCount: event.detail.roundCount }
			: undefined;
		try {
			const result = await startGame(selectedRegion || '', opts);
			sessionId = result.sessionId;
			currentQuestion = result.question;
			gameStarted = true;
			selectedAnswer = null;
			showFeedback = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to start game';
			console.error('Start game error:', err);
		} finally {
			isLoading = false;
		}
	}

	async function handleSelectAnswer(event: CustomEvent<{ country?: Country }>) {
		if (showFeedback || !sessionId || !currentQuestion || !event.detail.country) return;
		
		const country = event.detail.country;
		selectedAnswer = country.cca2;
		isLoading = true;
		error = null;
		
		try {
			const result = await submitAnswer(sessionId, currentQuestion.questionId, country.cca2);
			isCorrect = result.correct;
			// Translate the correct answer - find the country in options by CCA2 and translate it
			const correctCountry = currentQuestion.options.find(c => c.cca2 === result.correctCca2);
			correctAnswer = correctCountry ? getCountryNameForLocale(correctCountry) : result.correctName;
			score = result.score;
			total = result.total;
			showFeedback = true;
			
			setTimeout(async () => {
				showFeedback = false;
				if (result.finished) {
					gameFinished = true;
					gameStarted = false;
				} else {
					await loadNextQuestion();
				}
			}, 2000);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to submit answer';
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
			const question = await getQuestion(sessionId);
			currentQuestion = question;
			flagImageLoaded = false; // Reset flag loading state for new question
			// Reset selectedAnswer when new question loads
			selectedAnswer = null;
			// Focus input if using text input mode
			if (usesTextInput(difficultyMode) && guessInputElement) {
				setTimeout(() => {
					guessInputElement?.focus();
				}, 0);
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load question';
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

	// Re-fetch current question when locale changes (to update country names)
	// Initialize to null to avoid triggering on initial mount
	let previousLocale: string | null = null;
	$: if (gameStarted && currentQuestion && previousLocale !== null && currentLocale !== previousLocale && !showFeedback && !isLoading) {
		previousLocale = currentLocale;
		// Re-fetch the current question with new locale to get translated country names
		if (sessionId && currentQuestion) {
			getQuestion(sessionId).then(question => {
				currentQuestion = question;
				selectedAnswer = null;
				// Clear correctAnswer - it will be set again when user submits next answer
				correctAnswer = '';
			}).catch(err => {
				console.error('Failed to reload question with new locale:', err);
			});
		}
	}
	// Set previousLocale when game starts to track future locale changes
	$: if (gameStarted && currentQuestion && previousLocale === null) {
		previousLocale = currentLocale;
	}

	async function handleSubmitTextGuess() {
		if (showFeedback || !sessionId || !currentQuestion || !guessInput.trim() || isLoading) return;

		const countryName = guessInput.trim();
		isLoading = true;
		error = null;

		try {
			// Find country by name using reusable utility
			const matchedCountry = findCountryByName(allCountries, countryName, currentLocale);

			if (!matchedCountry) {
				error = t('game.guessing.not_found', undefined, currentLocale) || 'Country not found';
				isLoading = false;
				setTimeout(() => {
					guessInputElement?.focus();
				}, 0);
				return;
			}

			const result = await submitAnswer(sessionId, currentQuestion.questionId, matchedCountry.cca2);
			isCorrect = result.correct;
			const correctCountry = allCountries.find(c => c.cca2 === result.correctCca2);
			correctAnswer = correctCountry ? getCountryNameForLocale(correctCountry) : result.correctName;
			score = result.score;
			total = result.total;
			showFeedback = true;
			guessInput = '';

			setTimeout(async () => {
				showFeedback = false;
				if (result.finished) {
					gameFinished = true;
					gameStarted = false;
				} else {
					await loadNextQuestion();
				}
			}, 2000);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to submit answer';
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
	<title>{flagSeoTitle} - Flagged It</title>
	<meta name="description" content={flagPageDescription} />
</svelte:head>

<div class="min-h-screen p-4 md:p-8">
	<div class="max-w-4xl mx-auto">
		{#if !gameStarted && !gameFinished}
			<GameSetupScreen
				title={flagGameTitle}
				emoji="🚩"
				description={flagGameDescription}
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
			
			<GameContainer
				question={flagQuestionText}
				{showFeedback}
				{isCorrect}
				{correctAnswer}
				{error}
			>
				<div slot="content" class="relative w-full max-w-80 aspect-square flex items-center justify-center mx-auto">
					{#if !flagImageLoaded}
						<div class="absolute w-full max-w-80 aspect-[3/2.1] bg-white/5 rounded-[25px] animate-pulse border border-white/10"></div>
					{/if}
					<img 
						src={currentQuestion.flagUrl} 
						alt="Country flag" 
						class="w-full max-w-80 h-auto relative {flagImageLoaded ? 'opacity-100' : 'opacity-0'} transition-opacity duration-200"
						on:load={() => flagImageLoaded = true}
						on:error={() => flagImageLoaded = true}
					/>
				</div>
				
				<div slot="answers">
					{#if usesTextInput(difficultyMode)}
						<!-- Text input for expert mode -->
						<div class="w-full">
							<div class="flex flex-col sm:flex-row gap-3">
								<input
									type="text"
									bind:this={guessInputElement}
									bind:value={guessInput}
									on:keypress={handleKeyPress}
									placeholder={t('game.guessing.enter_country', undefined, currentLocale)}
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
						</div>
					{:else}
						<!-- Multiple choice for regular mode -->
						<AnswerGrid
							options={currentQuestion.options}
							{selectedAnswer}
							{correctAnswer}
							{showFeedback}
							{isCorrect}
							disabled={isLoading}
							on:select={handleSelectAnswer}
						/>
					{/if}
				</div>
			</GameContainer>
		{:else if isLoading}
			<div class="flex flex-col items-center justify-center py-20">
				<LoadingSpinner />
				<p class="mt-4 text-text-muted">Loading...</p>
			</div>
		{/if}
	</div>
</div>
