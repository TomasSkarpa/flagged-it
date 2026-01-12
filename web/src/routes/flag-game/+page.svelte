<script lang="ts">
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import { 
		GameSetupScreen, 
		GameHeader, 
		GameContainer, 
		AnswerGrid, 
		GameOverScreen 
	} from '$lib/components/game';
	import { startGame, getQuestion, submitAnswer } from '$lib/api/flagGame';
	import type { Country } from '$lib/types';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import { getCountryNameForLocale } from '$lib/utils/countryNames';
	import { calculateCurrentRound } from '$lib/utils/gameUtils';

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

	// Reactive translations - will update when locale changes
	$: currentLocale = $locale;
	$: flagGameTitle = t('game.flag.setup.title', undefined, currentLocale);
	$: flagGameDescription = t('game.flag.setup.description', undefined, currentLocale);
	$: flagQuestionText = t('game.flag.question', undefined, currentLocale);
	$: excellentMessage = t('game.flag.over.excellent', undefined, currentLocale);

	async function handleStartGame(event: CustomEvent<{ region: string }>) {
		isLoading = true;
		error = null;
		gameFinished = false;
		score = 0;
		total = 0;
		selectedRegion = event.detail.region;
		
		try {
			const result = await startGame(selectedRegion || '');
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

	async function handleSelectAnswer(event: CustomEvent<{ country: Country }>) {
		if (showFeedback || !sessionId || !currentQuestion) return;
		
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
		
		try {
			const question = await getQuestion(sessionId);
			currentQuestion = question;
			// Reset selectedAnswer when new question loads
			selectedAnswer = null;
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

	function handlePlayAgain() {
		gameStarted = false;
		gameFinished = false;
		currentQuestion = null;
		sessionId = null;
		score = 0;
		total = 0;
		selectedAnswer = null;
		showFeedback = false;
	}
</script>

<svelte:head>
	<title>Flag Game - Flagged It</title>
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
			/>
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
				<img 
					slot="content"
					src={currentQuestion.flagUrl} 
					alt="Country flag" 
					class="w-80 h-auto"
				/>
				
				<AnswerGrid
					slot="answers"
					key={currentQuestion.questionId}
					options={currentQuestion.options}
					{selectedAnswer}
					{correctAnswer}
					{showFeedback}
					{isCorrect}
					disabled={isLoading}
					on:select={handleSelectAnswer}
				/>
			</GameContainer>
		{:else if isLoading}
			<div class="flex flex-col items-center justify-center py-20">
				<LoadingSpinner />
				<p class="mt-4 text-text-muted">Loading...</p>
			</div>
		{/if}
	</div>
</div>
