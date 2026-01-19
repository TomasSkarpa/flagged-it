<script lang="ts">
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import ShapeRenderer from '$lib/components/ShapeRenderer.svelte';
	import MapShapeRenderer from '$lib/components/MapShapeRenderer.svelte';
	import { 
		GameSetupScreen, 
		GameHeader, 
		GameContainer, 
		AnswerGrid, 
		GameOverScreen,
		DifficultySelector
	} from '$lib/components/game';
	import type { DifficultyOption } from '$lib/components/game/DifficultySelector.svelte';
	import { startShapeGame, getShapeQuestion, submitShapeAnswer } from '$lib/api/shapeGame';
	import type { Country } from '$lib/types';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import { getCountryNameForLocale, findCountryByName } from '$lib/utils/countryNames';
	import { calculateCurrentRound } from '$lib/utils/gameUtils';
	import { getAllCountries } from '$lib/api/debug';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';

	type DifficultyMode = 'beginner' | 'intermediate' | 'advanced' | 'expert';

	// Reactive translations
	$: currentLocale = $locale;

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
	let difficultyMode: DifficultyMode = 'beginner';
	let gameStarted = false;
	let gameFinished = false;
	const totalRounds = 10;

	// Text input for intermediate/expert modes
	let guessInput = '';
	let guessInputElement: HTMLInputElement | null = null;
	let allCountries: Country[] = [];
	let countriesLoaded = false;

	// Reactive difficulty modes with translations
	let difficultyModes: DifficultyOption[] = [];
	$: difficultyModes = [
		{ 
			value: 'beginner', 
			label: t('game.shape.difficulty.beginner', undefined, currentLocale), 
			description: t('game.shape.difficulty.beginner.desc', undefined, currentLocale) 
		},
		{ 
			value: 'intermediate', 
			label: t('game.shape.difficulty.intermediate', undefined, currentLocale), 
			description: t('game.shape.difficulty.intermediate.desc', undefined, currentLocale) 
		},
		{ 
			value: 'advanced', 
			label: t('game.shape.difficulty.advanced', undefined, currentLocale), 
			description: t('game.shape.difficulty.advanced.desc', undefined, currentLocale) 
		},
		{ 
			value: 'expert', 
			label: t('game.shape.difficulty.expert', undefined, currentLocale), 
			description: t('game.shape.difficulty.expert.desc', undefined, currentLocale) 
		}
	];

	// Helper to check if mode uses text input
	function usesTextInput(mode: DifficultyMode): boolean {
		return mode === 'intermediate' || mode === 'expert';
	}

	// Helper to check if mode uses MapShapeRenderer
	function usesMapRenderer(mode: DifficultyMode): boolean {
		return mode === 'beginner' || mode === 'intermediate';
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

	async function handleStartGame(event: CustomEvent<{ region?: string; [key: string]: any }>) {
		isLoading = true;
		error = null;
		gameFinished = false;
		score = 0;
		total = 0;
		selectedRegion = event.detail.region || '';
		guessInput = '';
		
		try {
			const result = await startShapeGame(selectedRegion || '');
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

	async function handleSelectAnswer(event: CustomEvent<{ country?: Country; answer?: string; [key: string]: any }>) {
		if (showFeedback || !sessionId || !currentQuestion) return;
		
		const country = event.detail.country;
		if (!country) return;
		selectedAnswer = country.cca2;
		isLoading = true;
		error = null;
		
		try {
			const result = await submitShapeAnswer(sessionId, currentQuestion.questionId, country.cca2);
			isCorrect = result.correct;
			// Translate the correct answer - find the country in options by CCA2 and translate it
			const correctCountry = currentQuestion.options.find(c => c.cca2 === result.correctCca2);
			correctAnswer = correctCountry ? getCountryNameForLocale(correctCountry) : result.correctName;
			score = result.score;
			total = result.total;
			showFeedback = true;
			
			setTimeout(async () => {
				showFeedback = false;
				if (result.finished || total >= totalRounds) {
					gameFinished = true;
					gameStarted = false;
				} else {
					await loadNextQuestion();
				}
			}, 1200);
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
			const question = await getShapeQuestion(sessionId);
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
	let previousLocale: string = currentLocale;
	$: if (gameStarted && currentQuestion && currentLocale !== previousLocale && !showFeedback && !isLoading) {
		previousLocale = currentLocale;
		// Re-fetch the current question with new locale to get translated country names
		if (sessionId && currentQuestion) {
			getShapeQuestion(sessionId).then(question => {
				currentQuestion = question;
				selectedAnswer = null;
				// Clear correctAnswer - it will be set again when user submits next answer
				correctAnswer = '';
			}).catch(err => {
				console.error('Failed to reload question with new locale:', err);
			});
		}
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

			const result = await submitShapeAnswer(sessionId, currentQuestion.questionId, matchedCountry.cca2);
			isCorrect = result.correct;
			const correctCountry = allCountries.find(c => c.cca2 === result.correctCca2);
			correctAnswer = correctCountry ? getCountryNameForLocale(correctCountry) : result.correctName;
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
			}, 1200);
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
		difficultyMode = 'beginner'; // Reset to default
	}
</script>

<svelte:head>
	<title>Shape Game - Flagged It</title>
</svelte:head>

<div class="min-h-screen p-4 md:p-8">
	<div class="max-w-4xl mx-auto">
		{#if !gameStarted && !gameFinished}
			<GameSetupScreen
				title={t('game.shape.title')}
				emoji="🗺️"
				description={t('game.shape.description')}
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
				excellentMessage={t('game.over.excellent')}
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
				question={t('game.shape.question')}
				{showFeedback}
				{isCorrect}
				{correctAnswer}
				error={usesTextInput(difficultyMode) ? null : error}
			>
				<div slot="content" class="shape-renderer-container bg-white/5 rounded-lg p-4">
					{#if usesMapRenderer(difficultyMode)}
						<MapShapeRenderer 
							targetGeoJson={currentQuestion.geoJson}
							mode={difficultyMode === 'beginner' ? 'easy' : 'medium'}
							width={320}
							height={240}
							targetFillColor="var(--color-sandy-light)"
						/>
					{:else}
						<ShapeRenderer 
							geoJson={currentQuestion.geoJson}
							width={256}
							height={192}
							fillColor="var(--color-sandy-light)"
						/>
					{/if}
				</div>
				
				<div slot="answers">
					{#if usesTextInput(difficultyMode)}
						<!-- Text input for intermediate/expert modes -->
						<div class="card-game">
							<p class="text-center text-lg font-medium text-sandy-light mb-4">
								{t('game.guessing.make_guess', undefined, currentLocale)}
							</p>
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
							{#if error && !showFeedback}
								<div class="mt-3 p-3 bg-error/20 border border-error rounded-lg">
									<p class="text-error text-sm text-center font-medium">{error}</p>
								</div>
							{/if}
						</div>
					{:else}
						<!-- Multiple choice for beginner/advanced modes -->
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

<style>
	/* Shape renderer container - responsive sizing */
	.shape-renderer-container {
		display: flex;
		justify-content: center;
		align-items: center;
		width: 100%;
		max-width: 100%;
		overflow: hidden;
	}

	.shape-renderer-container :global(svg),
	.shape-renderer-container :global(.shape-renderer),
	.shape-renderer-container :global(.map-shape-renderer) {
		max-width: 100%;
		height: auto;
	}

	/* Responsive sizing for smaller screens */
	@media (max-width: 768px) {
		.shape-renderer-container :global(svg) {
			width: 100%;
			max-width: 480px;
		}
	}

	@media (max-width: 640px) {
		.shape-renderer-container {
			padding: 0.75rem;
		}
		
		.shape-renderer-container :global(svg) {
			max-width: 100%;
		}
	}

</style>
