<script lang="ts">
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import { onMount } from 'svelte';
	import {
		currentQuestion,
		currentRoom,
		currentPlayerId,
		submitAnswer,
		onMessage,
		wsConnected,
		wsConnection
	} from '$lib/stores/multiplayer';
	import { get } from 'svelte/store';
	import type { Question } from '$lib/stores/multiplayer';
	import type { Country } from '$lib/types';
	import AnswerGrid from '$lib/components/game/AnswerGrid.svelte';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';

	$: currentLocale = $locale;
	$: question = $currentQuestion;
	$: room = $currentRoom;
	$: questionIndex = room?.questionIndex || 0;
	$: totalQuestions = room?.config.numQuestions || 0;

	let questionStartTime = Date.now();
	let selectedAnswer: string | null = null;
	let showFeedback = false;
	let isCorrect = false;
	let correctAnswer = '';
	let timeRemaining = 0;

	// Reset timer when new question arrives
	$: if (question) {
		questionStartTime = Date.now();
		selectedAnswer = null;
		showFeedback = false;
		isCorrect = false;
		correctAnswer = '';
		timeRemaining = room?.config.timeLimit || 0;
	}

	// Update timer every second
	onMount(() => {
		const interval = setInterval(() => {
			if (room && room.config.timeLimit > 0 && question && !showFeedback) {
				const elapsed = Date.now() - questionStartTime;
				const remaining = Math.max(0, room.config.timeLimit * 1000 - elapsed);
				timeRemaining = Math.floor(remaining / 1000);
			}
		}, 1000); // Update every second

		return () => clearInterval(interval);
	});

	function handleSelectAnswer(event: CustomEvent<{ answer?: string; country?: Country }>) {
		console.log('handleSelectAnswer called:', { 
			hasQuestion: !!question, 
			playerId: $currentPlayerId, 
			showFeedback,
			selectedAnswer,
			wsConnected: $wsConnected,
			wsConnection: !!get(wsConnection),
			eventDetail: event.detail 
		});
		
		if (!question) {
			console.error('No question available');
			return;
		}
		
		if (!$currentPlayerId) {
			console.error('No player ID available');
			return;
		}
		
		// Prevent multiple submissions - if already answered, ignore
		if (showFeedback || selectedAnswer !== null) {
			console.log('Already answered or showing feedback, ignoring click');
			return;
		}
		
		// Check WebSocket connection
		if (!$wsConnected || !get(wsConnection)) {
			console.error('WebSocket not connected. State:', {
				wsConnected: $wsConnected,
				wsConnection: get(wsConnection),
				readyState: get(wsConnection)?.readyState
			});
			alert('Connection lost. Please refresh the page.');
			return;
		}
		
		// Extract answer value - could be string or Country CCA2
		let answer: string | undefined;
		if (event.detail.answer) {
			answer = event.detail.answer;
		} else if (event.detail.country) {
			answer = event.detail.country.cca2;
		}
		
		if (!answer) {
			console.error('No answer extracted from event:', event.detail);
			return;
		}

		console.log('Submitting answer:', { 
			playerId: $currentPlayerId, 
			questionId: question.id, 
			answer, 
			timeTaken: Date.now() - questionStartTime 
		});
		
		// Set selectedAnswer immediately to prevent multiple clicks
		selectedAnswer = answer;
		const timeTaken = Date.now() - questionStartTime;
		
		try {
			submitAnswer($currentPlayerId, question.id, answer, timeTaken);
			console.log('Answer submitted successfully');
		} catch (err) {
			console.error('Failed to submit answer:', err);
			// Reset selectedAnswer on error so user can try again
			selectedAnswer = null;
			alert('Failed to submit answer. Please check your connection.');
		}
	}

	// Listen for answer results
	onMount(() => {
		const unsubscribeAnswer = onMessage('ANSWER_RESULT', (msg) => {
			if (msg.type === 'ANSWER_RESULT' && msg.playerId === $currentPlayerId) {
				// Show feedback immediately
				showFeedback = true;
				isCorrect = msg.isCorrect;
				correctAnswer = msg.correctAnswer || '';
				
				// Auto-advance after 2 seconds (feedback will be reset when new question arrives)
				setTimeout(() => {
					showFeedback = false;
				}, 2000);
			}
		});

		// Also listen for ERROR messages (e.g., "already answered")
		const unsubscribeError = onMessage('ERROR', (msg) => {
			if (msg.type === 'ERROR' && msg.code === 'multiplayer.error.already_answered') {
				// If backend says already answered, reset selectedAnswer to allow retry
				// (though this shouldn't happen if frontend is working correctly)
				console.warn('Backend says already answered, resetting selectedAnswer');
				selectedAnswer = null;
				showFeedback = false;
			}
		});

		return () => {
			unsubscribeAnswer();
			unsubscribeError();
		};
	});

	// Render question based on game mode
	$: questionData = question?.data as Record<string, unknown> || {};
	$: questionType = question?.type || '';
	// Backend sends options as Country[] objects, not strings
	$: options = (question?.options || []) as Country[] | string[];
	$: flagUrl = questionData.flagUrl as string | undefined;
	$: shapeUrl = questionData.shapeUrl as string | undefined;
	$: questionText = (questionData.questionText as string) || 'Answer the question';
</script>

{#if question}
	<div class="card-game max-w-4xl mx-auto">
		<!-- Question Header -->
		<div class="flex items-center justify-between mb-6">
			<h2 class="text-xl md:text-2xl font-bold text-sandy-light">
				Question {questionIndex + 1} / {totalQuestions}
			</h2>
			{#if room && room.config.timeLimit > 0}
				<div class="text-lg font-semibold text-text-light">
					Time: <span id="timer">{timeRemaining}</span>s
				</div>
			{/if}
		</div>

		<!-- Question Content -->
		{#if questionType === 'flag'}
			<!-- Flag Question -->
			<div class="mb-8 flex justify-center">
				{#if flagUrl}
					<img 
						src={flagUrl} 
						alt="Country flag" 
						class="w-80 h-auto"
					/>
				{/if}
			</div>
			<div class="text-center mb-6">
				<h3 class="text-xl font-semibold text-text-light">
					{t('game.flag.question', undefined, currentLocale) || 'Which country is this flag from?'}
				</h3>
			</div>
		{:else if questionType === 'shape'}
			<!-- Shape Question -->
			<div class="mb-8 flex justify-center">
				{#if shapeUrl}
					<img 
						src={shapeUrl} 
						alt="Country shape" 
						class="w-80 h-auto"
					/>
				{/if}
			</div>
			<div class="text-center mb-6">
				<h3 class="text-xl font-semibold text-text-light">
					{t('game.shape.question', undefined, currentLocale) || 'Which country is this shape from?'}
				</h3>
			</div>
		{:else}
			<!-- Generic Question -->
			<div class="text-center mb-6">
				<h3 class="text-xl font-semibold text-text-light">
					{questionText}
				</h3>
			</div>
		{/if}

		<!-- Answer Options -->
		{#if options && options.length > 0}
			<div class="mb-4">
				<p class="text-sm text-text-muted mb-2">
					Debug: {options.length} options, Question: {question?.id}, Player: {$currentPlayerId || 'none'}
				</p>
			</div>
			<AnswerGrid
				{options}
				{selectedAnswer}
				{correctAnswer}
				{showFeedback}
				{isCorrect}
				disabled={showFeedback || selectedAnswer !== null}
				on:select={handleSelectAnswer}
			/>
		{:else}
			<div class="text-center text-text-muted">
				<p>No answer options available</p>
				{#if question}
					<pre class="text-xs mt-2 overflow-auto max-h-40">{JSON.stringify(question, null, 2)}</pre>
				{/if}
			</div>
		{/if}

		<!-- Feedback Overlay -->
		{#if showFeedback}
			<div 
				class="absolute inset-0 flex items-center justify-center rounded-card animate-fade-in z-20
					{isCorrect ? 'bg-success/50' : 'bg-error/50'}"
			>
				<div class="text-center">
					<div class="text-6xl mb-4">{isCorrect ? '✓' : '✗'}</div>
					<p class="text-3xl font-bold text-white">
						{isCorrect 
							? (t('game.correct_short', undefined, currentLocale) || 'Correct!')
							: (t('game.wrong_short', undefined, currentLocale) || 'Wrong!')}
					</p>
					{#if !isCorrect && correctAnswer}
						<p class="text-xl text-white/90 mt-2">{correctAnswer}</p>
					{/if}
				</div>
			</div>
		{/if}
	</div>
{:else}
	<div class="card-game text-center">
		<LoadingSpinner />
		<p class="mt-4 text-text-muted">Waiting for question...</p>
	</div>
{/if}
