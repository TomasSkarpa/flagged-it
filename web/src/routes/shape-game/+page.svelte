<script lang="ts">
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import ShapeRenderer from '$lib/components/ShapeRenderer.svelte';
	import { 
		GameSetupScreen, 
		GameHeader, 
		GameContainer, 
		AnswerGrid, 
		GameOverScreen 
	} from '$lib/components/game';
	import { startShapeGame, getShapeQuestion, submitShapeAnswer } from '$lib/api/shapeGame';
	import type { Country } from '$lib/types';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import { getCountryNameForLocale } from '$lib/utils/countryNames';
	import { calculateCurrentRound } from '$lib/utils/gameUtils';

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
	let gameStarted = false;
	let gameFinished = false;
	const totalRounds = 10;

	async function handleStartGame(event: CustomEvent<{ region: string }>) {
		isLoading = true;
		error = null;
		gameFinished = false;
		score = 0;
		total = 0;
		selectedRegion = event.detail.region;
		
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

	async function handleSelectAnswer(event: CustomEvent<{ country: Country }>) {
		if (showFeedback || !sessionId || !currentQuestion) return;
		
		const country = event.detail.country;
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
		
		try {
			const question = await getShapeQuestion(sessionId);
			currentQuestion = question;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load question';
			console.error('Load question error:', err);
		} finally {
			isLoading = false;
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
			/>
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
				{error}
			>
				<div slot="content" class="bg-white/5 rounded-lg p-4">
					<ShapeRenderer 
						geoJson={currentQuestion.geoJson}
						width={320}
						height={240}
						fillColor="var(--color-sandy-light)"
					/>
				</div>
				
				<AnswerGrid
					slot="answers"
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
