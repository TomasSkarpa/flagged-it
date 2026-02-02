<script lang="ts">
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import MapShapeRenderer from '$lib/components/MapShapeRenderer.svelte';
	import { 
		GameSetupScreen, 
		GameHeader, 
		GameContainer, 
		AnswerGrid, 
		GameOverScreen
	} from '$lib/components/game';
	// Timer will be implemented inline for now
	import { startSpeedChallenge, getQuestion, submitAnswer } from '$lib/api/speedChallengeGame';
	import type { SpeedChallengeQuestion, SpeedChallengeAnswer } from '$lib/api/speedChallengeGame';
	import type { Country } from '$lib/types';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import { getAllCountries } from '$lib/api/debug';
	import { calculateCurrentRound } from '$lib/utils/gameUtils';

	let sessionId: string | null = null;
	let currentQuestion: SpeedChallengeQuestion | null = null;
	let currentGameType: string = '';
	let selectedAnswer: string | null = null;
	let score = 0; // Total points
	let total = 0; // Rounds completed
	let correctAnswers = 0; // Number of correct answers
	let maxRounds = 10;
	let originalMaxRounds = 10; // Preserve original maxRounds for game over screen
	let timeLimit = 30; // seconds
	let timeRemaining = 30;
	let isLoading = false;
	let isSubmittingAnswer = false;
	let error: string | null = null;
	let showFeedback = false;
	let isCorrect = false;
	let correctAnswer = '';
	let gameStarted = false;
	let gameFinished = false;
	let questionStartTime = 0;
	let allCountries: Country[] = [];
	let countriesLoaded = false;
	let timerInterval: ReturnType<typeof setInterval> | null = null;

	// Initialize translation variables with defaults
	let speedChallengeTitle = 'Speed Challenge';
	let speedChallengeDescription = 'Test your geography knowledge across multiple game types with time pressure! Answer quickly to earn bonus points.';

	// Reactive translations with proper fallbacks
	$: currentLocale = $locale || 'en';
	$: {
		const title = t('game.speed_challenge.setup.title', undefined, currentLocale);
		if (title && typeof title === 'string') {
			speedChallengeTitle = title;
		}
		const desc = t('game.speed_challenge.setup.description', undefined, currentLocale);
		if (desc && typeof desc === 'string') {
			speedChallengeDescription = desc;
		}
	}

	onMount(async () => {
		if (!browser) return;
		try {
			const result = await getAllCountries();
			allCountries = result.countries;
			countriesLoaded = true;
		} catch (err) {
			console.warn('Failed to load countries:', err);
		}
	});

	onDestroy(() => {
		stopTimer();
	});

	function getGameTypeLabel(gameType: string): string {
		// Use .title keys (most games don't have .setup.title)
		const labels: Record<string, string> = {
			flag: t('game.flag.setup.title', undefined, currentLocale) || t('game.flag.title', undefined, currentLocale) || 'Flag Guessing Game',
			shape: t('game.shape.title', undefined, currentLocale) || 'Guess by Shape',
			capital: t('game.capital.title', undefined, currentLocale) || 'Capital City Game',
			facts: t('game.facts.title', undefined, currentLocale) || 'Guess by Facts',
			higher_lower: t('game.higher_lower.title', undefined, currentLocale) || 'Higher or Lower',
			worldle: t('game.worldle.title', undefined, currentLocale) || 'Worldle',
		};
		const result = labels[gameType] || gameType;
		// If result is still a translation key (contains dots), return a fallback
		if (result && typeof result === 'string' && result.includes('game.') && result.includes('.')) {
			// Fallback: capitalize and format the game type
			return gameType.split('_').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ');
		}
		return result || gameType;
	}

	function getQuestionText(gameType: string, question: SpeedChallengeQuestion | null): string {
		if (!question) return '';
		
		switch (gameType) {
			case 'flag':
				return t('game.flag.question', undefined, currentLocale) || 'Which country does this flag belong to?';
			case 'shape':
				return t('game.shape.question', undefined, currentLocale) || 'Which country is this?';
			case 'capital':
				return `What is the capital of ${question.countryName || 'this country'}?`;
			case 'facts':
				return t('game.facts.guess_country', undefined, currentLocale) || 'Guess the country based on the fact!';
			case 'higher_lower':
				return `Does ${question.right?.name || 'the second country'} have higher or lower ${question.valueLabel || 'population'} than ${question.left?.name || 'the first country'}?`;
			default:
				return 'Answer the question!';
		}
	}

	async function handleStartGame() {
		isLoading = true;
		error = null;
		gameFinished = false;
		score = 0;
		total = 0;
		correctAnswers = 0;
		timeRemaining = timeLimit;
		
		try {
			const result = await startSpeedChallenge(timeLimit, maxRounds);
			
			if (!result || !result.sessionId) {
				throw new Error('Invalid response from server');
			}
			
			sessionId = result.sessionId;
			
			// Check if question exists in response
			if (!result.question) {
				throw new Error('No question received from server');
			}
			
			currentQuestion = result.question;
			currentGameType = result.currentGameType || result.question?.gameType || '';
			maxRounds = result.maxRounds || 10;
			originalMaxRounds = maxRounds; // Preserve for game over screen
			timeLimit = result.timeLimit || 30;
			timeRemaining = timeLimit;
			score = result.score || 0;
			total = result.total || 0;
			
			gameStarted = true;
			selectedAnswer = null;
			showFeedback = false;
			correctAnswer = '';
			factsInput = '';
			startTimer();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to start game';
			console.error('Start game error:', err);
			gameStarted = false;
			currentQuestion = null;
		} finally {
			isLoading = false;
		}
	}

	function startTimer() {
		if (timerInterval) clearInterval(timerInterval);
		timeRemaining = timeLimit;
		questionStartTime = Date.now();
		
		timerInterval = setInterval(() => {
			if (showFeedback) return;
			
			const elapsed = (Date.now() - questionStartTime) / 1000;
			timeRemaining = Math.max(0, timeLimit - elapsed);
			
			if (timeRemaining <= 0) {
				clearInterval(timerInterval!);
				timerInterval = null;
				handleSubmitAnswer('');
			}
		}, 100);
	}

	function stopTimer() {
		if (timerInterval) {
			clearInterval(timerInterval);
			timerInterval = null;
		}
	}

	async function handleSubmitAnswer(answer: string) {
		if (showFeedback || isSubmittingAnswer || !sessionId || !currentQuestion) return;
		
		selectedAnswer = answer;
		isSubmittingAnswer = true;
		error = null;
		
		const timeTaken = Date.now() - questionStartTime;
		
		try {
			const result: SpeedChallengeAnswer = await submitAnswer(
				sessionId,
				currentQuestion.questionId,
				answer,
				timeTaken
			);
			
			// Update score immediately
			isCorrect = result.correct;
			correctAnswer = result.correctAnswer || '';
			score = result.score;
			total = result.total;
			// Only increment correctAnswers if answer is correct
			// Never set it to score or any other value - only increment
			if (result.correct === true) {
				correctAnswers = (correctAnswers || 0) + 1;
			}
			
			// Stop submitting state and show feedback overlay
			isSubmittingAnswer = false;
			showFeedback = true;
			
			// After feedback, transition to next question smoothly
			setTimeout(() => {
				showFeedback = false;
				
				// Small delay to allow feedback to fade out smoothly
				setTimeout(() => {
					if (result.finished) {
						// Ensure originalMaxRounds is preserved before finishing
						if (!originalMaxRounds || originalMaxRounds === 0) {
							originalMaxRounds = maxRounds || 10;
						}
						gameFinished = true;
						gameStarted = false;
					} else if (result.nextQuestion) {
						// Update question data smoothly
						currentQuestion = result.nextQuestion;
						currentGameType = result.nextGameType || '';
						timeLimit = result.timeLimit || timeLimit;
						selectedAnswer = null;
						correctAnswer = '';
						factsInput = '';
						startTimer();
					}
				}, 150);
			}, 2000);
		} catch (err) {
			isSubmittingAnswer = false;
			error = err instanceof Error ? err.message : 'Failed to submit answer';
			console.error('Submit answer error:', err);
		}
	}

	function handleAnswerSelect(event: CustomEvent<{ country?: Country }>) {
		if (showFeedback || !event.detail.country) return;
		handleSubmitAnswer(event.detail.country.cca2);
	}

	function handleCapitalSelect(capital: string) {
		if (showFeedback) return;
		handleSubmitAnswer(capital);
	}

	function handleHigherLowerSelect(direction: 'higher' | 'lower') {
		if (showFeedback) return;
		handleSubmitAnswer(direction);
	}

	function handleFactsSubmit() {
		if (showFeedback || !factsInput.trim()) return;
		handleSubmitAnswer(factsInput.trim());
	}

	let factsInput = '';
	let factsInputElement: HTMLInputElement | null = null;

	function handleFactsKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' && factsInput.trim()) {
			handleFactsSubmit();
		}
	}
</script>

<svelte:window on:keydown={handleFactsKeydown} />

<div class="min-h-screen p-4 md:p-8">
	<div class="max-w-4xl mx-auto">
		{#if !gameStarted && !gameFinished}
		<GameSetupScreen
			title={speedChallengeTitle || 'Speed Challenge'}
			description={speedChallengeDescription || 'Test your geography knowledge across multiple game types with time pressure!'}
			emoji="⚡"
			showRegionSelector={false}
			on:start={handleStartGame}
		/>
		{:else if gameFinished}
			{@const finalCorrectAnswers = Math.min(correctAnswers || 0, total || 0)}
			{@const finalMaxRounds = originalMaxRounds > 0 ? originalMaxRounds : (maxRounds > 0 ? maxRounds : 10)}
			<GameOverScreen
				score={finalCorrectAnswers}
				totalRounds={finalMaxRounds}
				on:playAgain={handleStartGame}
			/>
	{:else if gameStarted}
		{#if isLoading}
			<LoadingSpinner />
		{:else if error}
			<div class="error-container">
				<div class="error-message">{error}</div>
				<button class="btn-primary" on:click={handleStartGame}>Try Again</button>
			</div>
		{:else if currentQuestion}
			<GameHeader 
				score={correctAnswers} 
				total={total} 
				currentRound={calculateCurrentRound(total, maxRounds)}
				totalRounds={maxRounds}
			/>
			
			<div class="game-content-wrapper" class:fade-transition={showFeedback}>
				<GameContainer
					question={getQuestionText(currentGameType, currentQuestion)}
					showFeedback={showFeedback}
					{isCorrect}
					{correctAnswer}
					{error}
				>
					<!-- Speed Challenge specific: Timer and Game Type -->
					<div slot="header" class="mb-3 flex items-center justify-between w-full">
						<div class="text-sm text-text-muted">
							{getGameTypeLabel(currentGameType)}
						</div>
						<div class="timer-display">
							<span class="timer-value">{Math.floor(timeRemaining)}s</span>
						</div>
					</div>
					<div slot="content">
					{#if currentQuestion.gameType === 'flag'}
						<div class="relative w-full max-w-80 aspect-square flex items-center justify-center mx-auto">
							<img 
								src={currentQuestion.flagUrl} 
								alt="Country flag" 
								class="w-full max-w-80 h-auto"
							/>
						</div>
					{:else if currentQuestion.gameType === 'shape'}
						<div class="shape-renderer-container bg-white/5 rounded-lg p-4">
							{#if currentQuestion.geoJson}
								<MapShapeRenderer 
									targetGeoJson={currentQuestion.geoJson}
									mode="medium"
									width={320}
									height={240}
									targetFillColor="var(--color-sandy-light)"
								/>
							{/if}
						</div>
					{:else if currentQuestion.gameType === 'capital'}
						<!-- Capital question text is already in GameContainer question prop -->
					{:else if currentQuestion.gameType === 'facts'}
						<!-- Fact is shown in content slot, question prop is just the prompt -->
						<div class="w-full max-w-2xl mx-auto">
							<div class="bg-white/5 rounded-lg p-6 border border-white/10 min-h-[120px] flex items-center justify-center">
								<div class="text-center text-lg md:text-xl text-text-light leading-relaxed">
									{currentQuestion.fact}
								</div>
							</div>
						</div>
					{:else if currentQuestion.gameType === 'higher_lower'}
						<div class="higher-lower-comparison">
							<!-- Left Country -->
							<div class="comparison-panel comparison-left">
								{#if currentQuestion.left?.imageUrl}
									<img 
										src={currentQuestion.left.imageUrl} 
										alt="{currentQuestion.left?.name || ''} flag" 
										class="comparison-flag"
									/>
								{/if}
								<h3 class="comparison-country-name">"{currentQuestion.left?.name || 'Country 1'}"</h3>
								<p class="comparison-label">{t('game.higher_lower.has', undefined, currentLocale) || 'has'}</p>
								<p class="comparison-value">{currentQuestion.left?.value ? Number(currentQuestion.left.value).toLocaleString() : '?'}</p>
								<p class="comparison-category">{currentQuestion.valueLabel || 'population'}</p>
							</div>
							
							<!-- VS Circle -->
							<div class="vs-divider">
								<span>VS</span>
							</div>
							
							<!-- Right Country -->
							<div class="comparison-panel comparison-right">
								{#if currentQuestion.right?.imageUrl}
									<img 
										src={currentQuestion.right.imageUrl} 
										alt="{currentQuestion.right?.name || ''} flag" 
										class="comparison-flag"
									/>
								{/if}
								<h3 class="comparison-country-name">"{currentQuestion.right?.name || 'Country 2'}"</h3>
								<p class="comparison-label">{t('game.higher_lower.has', undefined, currentLocale) || 'has'}</p>
								<p class="comparison-value">?</p>
								<p class="comparison-category">{currentQuestion.valueLabel || 'population'}</p>
							</div>
						</div>
					{/if}
				</div>
				
				<div slot="answers">
					{#if currentQuestion.gameType === 'flag'}
						<AnswerGrid
							options={currentQuestion.options || []}
							{selectedAnswer}
							correctAnswer=""
							{showFeedback}
							{isCorrect}
							disabled={isSubmittingAnswer || showFeedback}
							on:select={handleAnswerSelect}
						/>
					{:else if currentQuestion.gameType === 'shape'}
						<AnswerGrid
							options={currentQuestion.options || []}
							{selectedAnswer}
							correctAnswer=""
							{showFeedback}
							{isCorrect}
							disabled={isSubmittingAnswer || showFeedback}
							on:select={handleAnswerSelect}
						/>
					{:else if currentQuestion.gameType === 'capital'}
						<div class="capital-options">
							{#if currentQuestion.options && Array.isArray(currentQuestion.options)}
								{#each currentQuestion.options as capital}
									<button
										class="capital-option"
										on:click={() => handleCapitalSelect(typeof capital === 'string' ? capital : String(capital))}
										disabled={showFeedback || isSubmittingAnswer}
									>
										{typeof capital === 'string' ? capital : String(capital)}
									</button>
								{/each}
							{:else}
								<p class="error-message">No options available</p>
							{/if}
						</div>
					{:else if currentQuestion.gameType === 'facts'}
						<div class="w-full">
							<div class="flex flex-col sm:flex-row gap-3">
								<input
									bind:this={factsInputElement}
									bind:value={factsInput}
									type="text"
									placeholder={t('game.guessing.enter_country', undefined, currentLocale) || 'Enter country name'}
									disabled={isSubmittingAnswer || showFeedback}
									class="flex-1 px-4 py-3 rounded-lg border-2 border-white/20 bg-white/5 text-sandy-light placeholder:text-text-muted focus:outline-none focus:border-primary transition-all disabled:opacity-50"
									autocomplete="off"
								/>
								<button
									on:click={handleFactsSubmit}
									disabled={isSubmittingAnswer || !factsInput.trim() || showFeedback}
									class="btn-primary px-8 py-3 font-semibold disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap"
								>
									{t('game.guessing.guess', undefined, currentLocale) || 'Submit'}
								</button>
							</div>
						</div>
					{:else if currentQuestion.gameType === 'higher_lower'}
						<div class="higher-lower-buttons">
							<button
								class="answer-btn higher"
								on:click={() => handleHigherLowerSelect('higher')}
								disabled={showFeedback || isSubmittingAnswer}
							>
								<span>{t('game.higher_lower.higher', undefined, currentLocale) || 'Higher'}</span>
								<span class="arrow">▲</span>
							</button>
							<button
								class="answer-btn lower"
								on:click={() => handleHigherLowerSelect('lower')}
								disabled={showFeedback || isSubmittingAnswer}
							>
								<span>{t('game.higher_lower.lower', undefined, currentLocale) || 'Lower'}</span>
								<span class="arrow">▼</span>
							</button>
						</div>
					{:else}
						<div class="error-message">
							Unknown game type: {currentQuestion.gameType || 'undefined'}
							<pre>{JSON.stringify(currentQuestion, null, 2)}</pre>
						</div>
					{/if}
				</div>
			</GameContainer>
			</div>
		{:else}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Loading question...</p>
			</div>
		{/if}
	{/if}
	</div>
</div>

<style>
	.timer-display {
		padding: 0.5rem 1rem;
		background: rgba(255, 255, 255, 0.1);
		border-radius: 0.5rem;
		border: 2px solid var(--color-primary);
		font-size: 1rem;
		font-weight: bold;
	}

	.timer-value {
		color: var(--color-primary);
		font-weight: bold;
	}

	.game-content-wrapper {
		transition: opacity 0.2s ease-in-out;
	}

	.game-content-wrapper.fade-transition {
		opacity: 0.7;
	}

	.capital-options {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1rem;
		max-width: 500px;
		margin: 0 auto;
	}

	.capital-option {
		padding: 1rem 2rem;
		font-size: 1.125rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
		background: rgba(255, 255, 255, 0.05);
		border-radius: 0.5rem;
		color: var(--color-text);
		cursor: pointer;
		transition: all 0.2s;
	}

	.capital-option:hover:not(:disabled) {
		background: rgba(255, 255, 255, 0.1);
		border-color: rgba(255, 255, 255, 0.2);
	}

	.capital-option:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.higher-lower-comparison {
		display: flex;
		align-items: stretch;
		gap: 2rem;
		width: 100%;
		max-width: 900px;
		margin: 0 auto;
		min-height: 400px;
	}

	.comparison-panel {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 2rem;
		border-radius: 1rem;
		background: rgba(255, 255, 255, 0.05);
		border: 2px solid rgba(255, 255, 255, 0.1);
		position: relative;
	}

	.comparison-left {
		background: linear-gradient(135deg, rgba(99, 102, 241, 0.2) 0%, rgba(139, 92, 246, 0.1) 100%);
	}

	.comparison-right {
		background: linear-gradient(135deg, rgba(236, 72, 153, 0.2) 0%, rgba(251, 113, 133, 0.1) 100%);
	}

	.comparison-flag {
		width: 120px;
		height: 90px;
		object-fit: contain;
		margin-bottom: 1rem;
		border-radius: 0.5rem;
		background: rgba(255, 255, 255, 0.1);
		padding: 0.5rem;
	}

	.comparison-country-name {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--color-text-light);
		margin: 0.5rem 0;
		text-align: center;
	}

	.comparison-label {
		font-size: 0.875rem;
		color: var(--color-text-muted);
		margin: 0.25rem 0;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.comparison-value {
		font-size: 2.5rem;
		font-weight: 800;
		color: var(--color-primary);
		margin: 0.5rem 0;
		line-height: 1;
	}

	.comparison-category {
		font-size: 1rem;
		color: var(--color-text-muted);
		margin-top: 0.5rem;
		text-transform: capitalize;
	}

	.vs-divider {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 60px;
		height: 60px;
		border-radius: 50%;
		background: var(--color-primary);
		color: white;
		font-size: 1.25rem;
		font-weight: 800;
		position: relative;
		z-index: 10;
		flex-shrink: 0;
		box-shadow: 0 4px 12px rgba(99, 102, 241, 0.4);
	}

	.higher-lower-buttons {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		align-items: center;
		margin-top: 1rem;
	}

	.answer-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.75rem;
		padding: 1rem 2rem;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-radius: 2rem;
		background: rgba(255, 255, 255, 0.1);
		color: var(--color-text-light);
		font-size: 1.125rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
		min-width: 180px;
	}

	.answer-btn:hover:not(:disabled) {
		background: rgba(255, 255, 255, 0.2);
		border-color: rgba(255, 255, 255, 0.5);
		transform: scale(1.05);
	}

	.answer-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.answer-btn .arrow {
		font-size: 0.875rem;
	}

	.error-message {
		color: var(--color-error);
		padding: 1rem;
		text-align: center;
	}

	.error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 50vh;
		gap: 2rem;
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 50vh;
		gap: 1rem;
	}

	.loading-container p {
		color: var(--color-text-muted);
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}
</style>
