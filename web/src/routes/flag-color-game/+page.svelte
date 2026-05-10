<script lang="ts">
	import { browser } from '$app/environment';
	import { onDestroy } from 'svelte';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import {
		GameSetupScreen,
		GameHeader,
		GameOverScreen,
		DifficultySelector,
	} from '$lib/components/game';
	import type { DifficultyOption } from '$lib/components/game/DifficultySelector.svelte';
	import FlagColorFlag from '$lib/components/game/FlagColorFlag.svelte';
	import FlagColorSvPlanePicker from '$lib/components/game/FlagColorSvPlanePicker.svelte';
	import FlagColorRoundResult from '$lib/components/game/FlagColorRoundResult.svelte';
	import {
		startFlagColorGame,
		getFlagColorQuestion,
		submitFlagColorAnswer,
	} from '$lib/api/flagColorGame';
	import type { FlagColorQuestion } from '$lib/api/flagColorGame';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import { calculateCurrentRound } from '$lib/utils/gameUtils';
	import { hsvToRgb, hexToRgb, rgbToHsv, formatHsbLabel } from '$lib/utils/colorHsv';
	import { flagColorTierFromRawDeltaE, pointsToTenDisplay } from '$lib/utils/flagColorScore';
	import { recordFlagColorRound } from '$lib/utils/flagColorStats';

	type DifficultyMode = 'easy' | 'hard';
	type Phase = 'playing' | 'roundResult';

	let sessionId: string | null = null;
	let currentQuestion: FlagColorQuestion | null = null;
	let score = 0;
	let total = 0;
	let isLoading = false;
	let error: string | null = null;
	let selectedRegion = '';
	let gameStarted = false;
	let gameFinished = false;
	const totalRounds = 10;

	let difficultyMode: DifficultyMode = 'easy';
	let hue = 200;
	let satPct = 85;
	let valPct = 75;
	$: sat = satPct / 100;
	$: val = valPct / 100;
	let maxPointsPerRound = 100;

	let phase: Phase = 'playing';
	let pendingFinished = false;
	let advanceTimer: ReturnType<typeof setTimeout> | null = null;

	let resultScoreTen = '0.00';
	let resultGuessHsb = '';
	let resultCorrectHsb = '';
	let resultGuessHex = '';
	let resultCorrectHex = '';
	let resultDeltaE = 0;
	let resultTierMessage = '';

	$: currentLocale = $locale;
	$: rgbPreview = hsvToRgb(hue, sat, val);

	function clearAdvanceTimer(): void {
		if (advanceTimer !== null) {
			clearTimeout(advanceTimer);
			advanceTimer = null;
		}
	}

	onDestroy(() => clearAdvanceTimer());

	$: difficultyModes = [
		{
			value: 'easy',
			label: t('game.flag_color.difficulty.easy', undefined, currentLocale),
			description: t('game.flag_color.difficulty.easy.desc', undefined, currentLocale),
		},
		{
			value: 'hard',
			label: t('game.flag_color.difficulty.hard', undefined, currentLocale),
			description: t('game.flag_color.difficulty.hard.desc', undefined, currentLocale),
		},
	] satisfies DifficultyOption[];

	$: scoreDenominator = Math.min(
		totalRounds * maxPointsPerRound,
		(total + (phase === 'playing' && gameStarted && !gameFinished ? 1 : 0)) * maxPointsPerRound
	);

	$: setupTitle = t('game.flag_color.setup.title', undefined, currentLocale);
	$: setupDescription = t('game.flag_color.setup.description', undefined, currentLocale);
	$: pageTitle = t('game.flag_color.title', undefined, currentLocale);
	$: pageDesc = t('home.game.flag_color.description', undefined, currentLocale);
	$: questionBeforeCountry = t(
		'game.flag_color.question.before_country',
		undefined,
		currentLocale
	);
	$: questionAfterCountry = t(
		'game.flag_color.question.after_country',
		undefined,
		currentLocale
	);
	$: labelYourSelection = t('game.flag_color.result.your_selection', undefined, currentLocale);
	$: labelOriginal = t('game.flag_color.result.original', undefined, currentLocale);
	$: hueAriaLabel = t('game.flag_color.hue', undefined, currentLocale);

	async function handleStartGame(event: CustomEvent<{ region?: string }>) {
		clearAdvanceTimer();
		isLoading = true;
		error = null;
		gameFinished = false;
		score = 0;
		total = 0;
		phase = 'playing';
		selectedRegion = event.detail.region || '';
		try {
			const result = await startFlagColorGame(selectedRegion || '', {
				roundCount: totalRounds,
				difficulty: difficultyMode,
			});
			sessionId = result.sessionId;
			currentQuestion = result.question;
			maxPointsPerRound = result.question.maxPointsRound ?? 100;
			gameStarted = true;
			hue = 200;
			satPct = 85;
			valPct = 75;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to start';
		} finally {
			isLoading = false;
		}
	}

	async function loadNextQuestion() {
		if (!sessionId) return;
		isLoading = true;
		error = null;
		try {
			const question = await getFlagColorQuestion(sessionId);
			currentQuestion = question;
			maxPointsPerRound = question.maxPointsRound ?? maxPointsPerRound;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load question';
		} finally {
			isLoading = false;
		}
	}

	async function handleRoundContinue() {
		clearAdvanceTimer();
		if (pendingFinished) {
			phase = 'playing';
			gameFinished = true;
			gameStarted = false;
			pendingFinished = false;
			return;
		}
		phase = 'playing';
		resultScoreTen = '0.00';
		await loadNextQuestion();
		hue = 200;
		satPct = 85;
		valPct = 75;
	}

	async function handleSubmit() {
		if (phase !== 'playing' || !sessionId || !currentQuestion || isLoading) return;
		isLoading = true;
		error = null;
		const q = currentQuestion;
		try {
			const r = await submitFlagColorAnswer(sessionId, q.questionId, rgbPreview.r, rgbPreview.g, rgbPreview.b);
			score = r.score;
			total = r.total;
			resultScoreTen = pointsToTenDisplay(r.pointsEarned, r.maxPointsPerRound);
			resultTierMessage = flagColorTierFromRawDeltaE(r.deltaE, currentLocale);
			resultGuessHsb = formatHsbLabel(hue, sat, val);
			resultGuessHex = r.guessHex;
			resultCorrectHex = r.correctHex;
			resultDeltaE = r.deltaE;
			const cr = hexToRgb(r.correctHex);
			const chsv = rgbToHsv(cr.r, cr.g, cr.b);
			resultCorrectHsb = formatHsbLabel(chsv.h, chsv.s, chsv.v);

			if (browser) {
				recordFlagColorRound({
					cca2: q.cca2,
					deltaE: r.deltaE,
					pointsEarned: r.pointsEarned,
					difficulty: difficultyMode,
				});
			}

			pendingFinished = r.finished;
			phase = 'roundResult';
			clearAdvanceTimer();
			advanceTimer = setTimeout(() => {
				advanceTimer = null;
				void handleRoundContinue();
			}, 4200);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to submit';
		} finally {
			isLoading = false;
		}
	}

	function handleDifficultySelect(event: CustomEvent<{ value: string }>) {
		difficultyMode = event.detail.value as DifficultyMode;
	}

	function handlePlayAgain() {
		clearAdvanceTimer();
		gameStarted = false;
		gameFinished = false;
		currentQuestion = null;
		sessionId = null;
		score = 0;
		total = 0;
		phase = 'playing';
		difficultyMode = 'easy';
	}

	$: playingTint =
		gameStarted && !gameFinished && phase === 'playing'
			? `radial-gradient(ellipse 120% 80% at 50% -20%, rgba(${rgbPreview.r},${rgbPreview.g},${rgbPreview.b},0.22), transparent 50%)`
			: '';
</script>

<svelte:head>
	<title>{pageTitle} - Flagged It</title>
	<meta name="description" content={pageDesc} />
</svelte:head>

<div
	class="min-h-screen p-4 md:p-8 transition-[background] duration-300"
	style={playingTint ? `background-image: ${playingTint};` : ''}
>
	<div class="max-w-5xl mx-auto">
		{#if !gameStarted && !gameFinished}
			<GameSetupScreen
				title={setupTitle}
				emoji="🎨"
				description={setupDescription}
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
				scoreDenominator={totalRounds * maxPointsPerRound}
				excellentMessage={t('game.flag_color.over.excellent', undefined, currentLocale)}
				on:playAgain={handlePlayAgain}
			/>
		{:else if currentQuestion}
			<GameHeader
				{score}
				{total}
				currentRound={calculateCurrentRound(total, totalRounds)}
				{totalRounds}
				scoreDenominator={scoreDenominator > 0 ? scoreDenominator : maxPointsPerRound}
			/>

			<div class="card-game overflow-hidden">
				<h2
					class="flag-color-headline text-lg sm:text-xl md:text-2xl font-bold text-center mb-5 px-3 md:px-6 leading-snug max-w-3xl mx-auto"
				>
					<span class="text-sandy-light/95 font-semibold">{questionBeforeCountry}</span><span
						class="flag-color-headline-country">{currentQuestion.countryName}</span><span class="text-sandy-light/95 font-semibold">{questionAfterCountry}</span>
				</h2>

				{#if phase === 'playing'}
					<div class="flex flex-col gap-6 px-2 pb-4 min-w-0">
						<div
							class="flex flex-col lg:flex-row lg:items-start lg:justify-center gap-6 lg:gap-12 xl:gap-14 w-full min-w-0"
						>
							<div
								class="flex flex-col sm:flex-row lg:flex-col items-center justify-center gap-3 sm:gap-4 w-full min-w-0 lg:w-auto lg:max-w-[320px]"
							>
								<div class="w-full min-w-0 flex justify-center shrink-0">
									<FlagColorSvPlanePicker
										bind:hue
										bind:satPct
										bind:valPct
										disabled={isLoading}
										{hueAriaLabel}
										size="game"
									/>
								</div>
								<div
									class="hidden sm:block sm:w-28 sm:min-h-[168px] sm:shrink-0 lg:w-full lg:max-w-[300px] lg:min-h-0 lg:h-14 lg:shrink-0 rounded-2xl lg:rounded-xl border border-white/15 shadow-inner transition-colors duration-150"
									style="background-color: rgb({rgbPreview.r},{rgbPreview.g},{rgbPreview.b});"
									aria-hidden="true"
								></div>
							</div>
							<div class="flex justify-center shrink-0 lg:pt-1">
								<FlagColorFlag
									flagUrl={currentQuestion.flagUrl}
									guessableId={currentQuestion.guessableId}
									questionId={currentQuestion.questionId}
									revealFullColor={false}
									previewR={rgbPreview.r}
									previewG={rgbPreview.g}
									previewB={rgbPreview.b}
								/>
							</div>
						</div>
						<div class="flex justify-center w-full">
							<button
								type="button"
								class="btn-primary w-full max-w-md px-8 py-3.5 font-semibold rounded-2xl disabled:opacity-50 shadow-lg"
								disabled={isLoading}
								on:click={handleSubmit}
							>
								{t('game.flag_color.submit', undefined, currentLocale)}
							</button>
						</div>
					</div>
				{:else}
					<div class="flex flex-col md:flex-row gap-6 md:gap-8 items-start justify-center px-2 pb-6">
						<div class="shrink-0 mx-auto md:mx-0">
							<FlagColorFlag
								flagUrl={currentQuestion.flagUrl}
								guessableId={currentQuestion.guessableId}
								questionId={currentQuestion.questionId}
								revealFullColor={true}
								previewR={rgbPreview.r}
								previewG={rgbPreview.g}
								previewB={rgbPreview.b}
							/>
						</div>
						<div class="flex-1 w-full min-w-0">
							<FlagColorRoundResult
								roundCurrent={total}
								roundTotal={totalRounds}
								scoreTen={resultScoreTen}
								deltaE={resultDeltaE}
								tierMessage={resultTierMessage}
								guessHsb={resultGuessHsb}
								correctHsb={resultCorrectHsb}
								guessHex={resultGuessHex}
								correctHex={resultCorrectHex}
								{labelYourSelection}
								{labelOriginal}
								on:continue={() => void handleRoundContinue()}
							/>
						</div>
					</div>
				{/if}

				{#if error}
					<div class="mt-4 mx-2 mb-4 p-4 bg-error/20 border border-error rounded-xl">
						<p class="text-error font-semibold text-center">{error}</p>
					</div>
				{/if}
			</div>

			{#if isLoading && phase === 'playing'}
				<div class="flex justify-center mt-6">
					<LoadingSpinner />
				</div>
			{/if}
		{/if}
	</div>
</div>

<style>
	/* Accent styling for the country name in the headline (dark UI). */
	.flag-color-headline-country {
		display: inline;
		font-weight: 800;
		color: #e891c7;
		text-shadow:
			0 0 28px rgba(232, 145, 199, 0.45),
			0 1px 2px rgba(0, 0, 0, 0.35);
		margin-left: 0.06em;
		margin-right: 0.06em;
	}

	:global(:root.light) .flag-color-headline-country {
		color: #b83280;
		text-shadow: none;
	}
</style>
