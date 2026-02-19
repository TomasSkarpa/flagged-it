<script lang="ts">
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import {
		GameSetupScreen,
		GameHeader,
		GameContainer,
		AnswerGrid,
		GameOverScreen
	} from '$lib/components/game';
	import { startQuiz, submitQuizRound } from '$lib/api/quizGame';
	import type { QuizGameType } from '$lib/api/quizGame';
	import type { Country } from '$lib/types';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import { resolveAssetUrl } from '$lib/api/config';
	import { calculateCurrentRound } from '$lib/utils/gameUtils';

	let quizSessionId: string | null = null;
	let score = 0;
	let totalRounds = 0;
	let currentRound = 0;
	let gameType: string | null = null;
	let roundData: unknown = null;
	let isLoading = false;
	let error: string | null = null;
	let gameStarted = false;
	let gameComplete = false;
	let showFeedback = false;
	let isCorrect = false;
	let correctAnswer = '';
	let selectedAnswer: string | null = null;
	let revealedAnswer: unknown = null;
	let flagImageLoaded = false;

	$: flagQuestion =
		gameType === 'flag' && roundData
			? (roundData as { flagUrl: string; options: Country[]; questionId: string })
			: null;

	$: flagImgSrc = flagQuestion?.flagUrl ? resolveAssetUrl(flagQuestion.flagUrl) : '';
	$: currentLocale = $locale;

	/** Round types that have quiz UI implemented. Only these are offered and accepted. */
	const SUPPORTED_QUIZ_ROUND_TYPES: QuizGameType[] = ['flag'];

	$: presets = [
		{ label: t('game.quiz.preset.flag_only', undefined, currentLocale) || 'Flag only (5 rounds)', types: ['flag', 'flag', 'flag', 'flag', 'flag'] as QuizGameType[] }
	];

	$: quizTitle = t('game.quiz.title', undefined, currentLocale) || 'Quiz';
	$: quizDescription = t('game.quiz.description', undefined, currentLocale) || 'Play a mix of rounds from different games.';
	$: flagQuestionText = t('game.flag.question', undefined, currentLocale) || 'Which country is this?';
	$: excellentMessage = t('game.quiz.over.excellent', undefined, currentLocale) || t('game.over.excellent', undefined, currentLocale) || 'Excellent!';
	$: gameCompleteTitle = t('game.quiz.over.title', undefined, currentLocale) || 'Quiz complete';
	$: unsupportedRoundMessage = (() => {
		const msg = t('game.quiz.unsupported_round', undefined, currentLocale);
		return msg && msg !== 'game.quiz.unsupported_round' ? msg : 'This round type is not yet supported in quiz. Play the dedicated game for this mode.';
	})();

	async function handleStart(types: QuizGameType[]) {
		const supportedTypes = types.filter((type) => SUPPORTED_QUIZ_ROUND_TYPES.includes(type));
		if (supportedTypes.length === 0) {
			error = t('game.quiz.no_supported_rounds', undefined, currentLocale) || 'No supported round types in this preset.';
			return;
		}
		isLoading = true;
		error = null;
		gameComplete = false;
		flagImageLoaded = false;
		try {
			const result = await startQuiz(supportedTypes);
			quizSessionId = result.quizSessionId;
			score = result.score;
			totalRounds = result.totalRounds;
			currentRound = result.currentRound;
			gameType = result.gameType ?? null;
			roundData = result.data ?? null;
			gameStarted = true;
			showFeedback = false;
			revealedAnswer = null;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to start quiz';
		} finally {
			isLoading = false;
		}
	}

	async function handleSubmit(data: Record<string, unknown>) {
		if (!quizSessionId || isLoading) return;
		isLoading = true;
		error = null;
		showFeedback = false;
		revealedAnswer = null;
		try {
			const result = await submitQuizRound(quizSessionId, data);
			showFeedback = true;
			isCorrect = result.correct;
			score = result.score;
			currentRound = result.currentRound;
			revealedAnswer = result.revealedAnswer ?? null;
			correctAnswer =
				revealedAnswer && typeof revealedAnswer === 'object' && 'correctName' in revealedAnswer
					? String((revealedAnswer as { correctName?: string }).correctName ?? '')
					: '';
			if (result.complete) {
				gameComplete = true;
				return;
			}
			if (result.nextGameType != null) {
				gameType = result.nextGameType;
				roundData = result.nextData ?? null;
				flagImageLoaded = false;
				selectedAnswer = null;
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to submit';
		} finally {
			isLoading = false;
		}
	}

	function handleSelectOption(country: Country) {
		selectedAnswer = country.cca2;
		handleSubmit({ cca2: country.cca2 });
	}

	function handlePlayAgain() {
		gameStarted = false;
		gameComplete = false;
		quizSessionId = null;
		gameType = null;
		roundData = null;
		score = 0;
		currentRound = 0;
		showFeedback = false;
		revealedAnswer = null;
	}
</script>

<svelte:head>
	<title>Quiz | Flagged It</title>
</svelte:head>

<div class="min-h-screen p-4 md:p-8">
	<div class="max-w-4xl mx-auto">
		{#if !gameStarted}
			<GameSetupScreen
				title={quizTitle}
				emoji="🎲"
				description={quizDescription}
				{isLoading}
				{error}
				showRegionSelector={false}
				showStartButton={false}
			>
				<div slot="options" class="mb-8">
					<div class="max-w-2xl mx-auto">
						{#if error}
							<p class="text-error font-semibold mb-4">{error}</p>
						{/if}
						<div class="flex flex-col gap-3">
							{#each presets as preset}
								<button
									class="btn-primary w-full px-6 py-4 text-lg font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
									disabled={isLoading}
									on:click={() => handleStart(preset.types)}
								>
									{preset.label}
								</button>
							{/each}
						</div>
						{#if isLoading}
							<div class="flex flex-col items-center justify-center py-8">
								<LoadingSpinner />
								<p class="mt-4 text-text-muted">{t('game.setup.starting', undefined, currentLocale) || 'Starting...'}</p>
							</div>
						{/if}
					</div>
				</div>
			</GameSetupScreen>
		{:else if gameComplete}
			<GameOverScreen
				{score}
				{totalRounds}
				title={gameCompleteTitle}
				excellentMessage={excellentMessage}
				on:playAgain={handlePlayAgain}
			/>
		{:else if flagQuestion}
			<GameHeader
				{score}
				total={totalRounds}
				currentRound={calculateCurrentRound(score, totalRounds)}
				{totalRounds}
			/>

			<GameContainer
				question={flagQuestionText}
				{showFeedback}
				{isCorrect}
				correctAnswer={correctAnswer}
				{error}
			>
				<div slot="content" class="relative w-full max-w-80 aspect-square flex items-center justify-center mx-auto">
					{#if !flagImageLoaded}
						<div class="absolute w-full max-w-80 aspect-[3/2.1] bg-white/5 rounded-[25px] animate-pulse border border-white/10"></div>
					{/if}
					<img
						src={flagImgSrc}
						alt="Country flag"
						class="w-full max-w-80 h-auto relative {flagImageLoaded ? 'opacity-100' : 'opacity-0'} transition-opacity duration-200"
						on:load={() => (flagImageLoaded = true)}
						on:error={() => (flagImageLoaded = true)}
					/>
				</div>

				<div slot="answers">
					<AnswerGrid
						options={flagQuestion.options}
						{selectedAnswer}
						correctAnswer={showFeedback ? correctAnswer : null}
						{showFeedback}
						{isCorrect}
						disabled={isLoading}
						on:select={(e) => e.detail?.country && handleSelectOption(e.detail.country)}
					/>
				</div>
			</GameContainer>
		{:else if gameType}
			<GameHeader
				{score}
				total={totalRounds}
				currentRound={calculateCurrentRound(score, totalRounds)}
				{totalRounds}
			/>
			<GameContainer question={gameType ? `Round: ${gameType}` : 'Round'} {error}>
				<div slot="content" class="p-6 bg-white/5 rounded-lg border border-white/10 text-center">
					<p class="text-text-muted">{unsupportedRoundMessage}</p>
				</div>
			</GameContainer>
		{:else if isLoading}
			<div class="flex flex-col items-center justify-center py-20">
				<LoadingSpinner />
				<p class="mt-4 text-text-muted">{t('game.setup.starting', undefined, currentLocale) || 'Loading...'}</p>
			</div>
		{/if}
	</div>
</div>
