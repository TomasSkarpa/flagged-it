<script lang="ts">
	import { onMount } from 'svelte';
	import { 
		startHigherLowerGame, 
		submitHigherLowerAnswer,
		formatValue,
		getCategoryLabel,
		getCategoryDescription,
		type HigherLowerCategory,
		type HigherLowerComparison,
		type HigherLowerItem
	} from '$lib/api/higherLowerGame';
	import { 
		GameSetupScreen, 
		GameOverScreen 
	} from '$lib/components/game';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';

	let sessionId: string | null = null;
	let comparison: HigherLowerComparison | null = null;
	let score = 0;
	let highScore = 0;
	let isLoading = false;
	let error: string | null = null;
	let gameStarted = false;
	let gameOver = false;
	let selectedCategory: HigherLowerCategory = 'population';
	
	// Animation states
	let showRightValue = false;
	let isCorrect: boolean | null = null;
	let isAnimating = false;
	let slideDirection: 'left' | null = null;

	// Reactive translations - will update when locale changes
	$: currentLocale = $locale;
	$: setupDescription = t('game.higher_lower.setup.description', undefined, currentLocale);
	$: higherLowerTitle = t('game.higher_lower.title', undefined, currentLocale);
	$: gameOverText = t('game.higher_lower.game_over', undefined, currentLocale);
	$: finalScoreText = t('game.higher_lower.final_score', undefined, currentLocale);
	$: highScoreText = t('game.higher_lower.high_score', undefined, currentLocale);
	$: loadingText = t('game.higher_lower.loading', undefined, currentLocale);
	$: startText = t('game.higher_lower.start', undefined, currentLocale);
	$: playAgainText = t('game.over.play_again', undefined, currentLocale);
	$: hasText = t('game.higher_lower.has', undefined, currentLocale);
	$: thanText = t('game.higher_lower.than', undefined, currentLocale);
	$: higherText = t('game.higher_lower.higher', undefined, currentLocale);
	$: lowerText = t('game.higher_lower.lower', undefined, currentLocale);
	$: scoreLabelText = t('game.higher_lower.score', [score], currentLocale);

	$: categories = [
		{ value: 'population' as HigherLowerCategory, label: t('game.higher_lower.category.population', undefined, currentLocale), icon: '👥', description: t('game.higher_lower.category.population.desc', undefined, currentLocale) },
		{ value: 'area' as HigherLowerCategory, label: t('game.higher_lower.category.area', undefined, currentLocale), icon: '📐', description: t('game.higher_lower.category.area.desc', undefined, currentLocale) },
		{ value: 'continents' as HigherLowerCategory, label: t('game.higher_lower.category.continents', undefined, currentLocale), icon: '🌍', description: t('game.higher_lower.category.continents.desc', undefined, currentLocale) }
	];
	
	// Custom start data that will be passed with the start event (reactive)
	$: customStartData = { category: selectedCategory };

	async function handleStartGame(event: CustomEvent<{ category?: string; region?: string; [key: string]: any }>) {
		// Get category from event or fallback to selectedCategory
		const category = (event.detail.category || selectedCategory) as HigherLowerCategory;
		isLoading = true;
		error = null;
		gameOver = false;
		showRightValue = false;
		isCorrect = null;
		
		try {
			const result = await startHigherLowerGame(category);
			sessionId = result.sessionId;
			comparison = result.comparison;
			score = result.score;
			highScore = result.highScore;
			gameStarted = true;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to start game';
		} finally {
			isLoading = false;
		}
	}

	async function handleAnswer(answer: 'higher' | 'lower') {
		if (!sessionId || isAnimating) return;
		
		isAnimating = true;
		showRightValue = true;
		
		try {
			const result = await submitHigherLowerAnswer(sessionId, answer);
			isCorrect = result.correct;
			score = result.score;
			highScore = result.highScore;
			
			if (result.gameOver) {
				setTimeout(() => {
					gameOver = true;
					isAnimating = false;
				}, 2000);
			} else if (result.nextComparison) {
				// Wait for animation, then slide
				setTimeout(() => {
					slideDirection = 'left';
					setTimeout(() => {
						comparison = result.nextComparison!;
						showRightValue = false;
						isCorrect = null;
						slideDirection = null;
						isAnimating = false;
					}, 500);
				}, 1500);
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to submit answer';
			isAnimating = false;
		}
	}

	function handlePlayAgain() {
		gameStarted = false;
		gameOver = false;
		comparison = null;
		sessionId = null;
		score = 0;
		showRightValue = false;
		isCorrect = null;
	}

	function getBackgroundGradient(item: HigherLowerItem, side: 'left' | 'right'): string {
		// Generate unique gradients based on country/continent name
		const hash = item.name.split('').reduce((a, b) => {
			a = ((a << 5) - a) + b.charCodeAt(0);
			return a & a;
		}, 0);
		
		const hue1 = Math.abs(hash % 360);
		const hue2 = (hue1 + 40) % 360;
		
		if (side === 'left') {
			return `linear-gradient(135deg, hsl(${hue1}, 60%, 25%) 0%, hsl(${hue2}, 50%, 15%) 100%)`;
		}
		return `linear-gradient(135deg, hsl(${hue2}, 60%, 25%) 0%, hsl(${hue1}, 50%, 15%) 100%)`;
	}
</script>

<svelte:head>
	<title>Higher or Lower - Flagged It</title>
</svelte:head>

{#if !gameStarted && !gameOver}
	<div class="min-h-screen p-4 md:p-8">
		<div class="max-w-4xl mx-auto">
			<GameSetupScreen
				title={higherLowerTitle}
				emoji="↕️"
				description={setupDescription}
				{isLoading}
				{error}
				showRegionSelector={false}
				startButtonText={startText}
				loadingText={loadingText}
				{customStartData}
				on:start={handleStartGame}
			>
				<div slot="options" class="mb-8">
					<div class="category-grid max-w-2xl mx-auto">
						{#each categories as cat}
							<button
								class="category-card"
								class:selected={selectedCategory === cat.value}
								on:click={() => {
									selectedCategory = cat.value;
								}}
								type="button"
							>
								<span class="category-icon">{cat.icon}</span>
								<span class="category-label">{cat.label}</span>
								<span class="category-desc">{cat.description}</span>
							</button>
						{/each}
					</div>
				</div>
			</GameSetupScreen>
		</div>
	</div>
{:else if gameOver}
	<div class="min-h-screen p-4 md:p-8">
		<div class="max-w-4xl mx-auto">
			<!-- Game Over Screen - Custom for Higher/Lower -->
			<div class="text-center">
				<div class="card-game max-w-2xl mx-auto">
					<div class="text-6xl mb-6">🎯</div>
					<h2 class="text-4xl font-bold text-sandy-light mb-4">{gameOverText}</h2>
					<div class="score-display mb-8">
						<div class="final-score">
							<span class="score-label">{finalScoreText}</span>
							<span class="score-value">{score}</span>
						</div>
						<div class="high-score">
							<span class="score-label">{highScoreText}</span>
							<span class="score-value">{highScore}</span>
						</div>
					</div>
					<button
						class="btn-primary px-12 py-4 text-xl font-bold"
						on:click={handlePlayAgain}
					>
						{playAgainText}
					</button>
				</div>
			</div>
		</div>
	</div>
{:else if comparison}
	<!-- Game Screen - Split View (full screen) -->
	<div class="game-screen" class:slide-left={slideDirection === 'left'}>
			<!-- Left Panel -->
			<div 
				class="panel panel-left"
				style="background: {getBackgroundGradient(comparison.left, 'left')}"
			>
				<div class="panel-content">
					{#if comparison.left.cca2}
						<img 
							src={comparison.left.imageUrl} 
							alt="{comparison.left.name} flag" 
							class="panel-image"
						/>
					{:else}
						<span class="panel-emoji">{comparison.left.imageUrl}</span>
					{/if}
					
					<h2 class="panel-title">"{comparison.left.name}"</h2>
					<p class="panel-subtitle">{hasText}</p>
					<p class="panel-value">{formatValue(comparison.left.value, comparison.category)}</p>
					<p class="panel-label">{comparison.valueLabel}</p>
				</div>
			</div>
			
			<!-- VS Circle -->
			<div class="vs-circle">
				<span>VS</span>
			</div>
			
			<!-- Right Panel -->
			<div 
				class="panel panel-right"
				style="background: {getBackgroundGradient(comparison.right, 'right')}"
			>
				<div class="panel-content">
					{#if comparison.right.cca2}
						<img 
							src={comparison.right.imageUrl} 
							alt="{comparison.right.name} flag" 
							class="panel-image"
						/>
					{:else}
						<span class="panel-emoji">{comparison.right.imageUrl}</span>
					{/if}
					
					<h2 class="panel-title">"{comparison.right.name}"</h2>
					<p class="panel-subtitle">{hasText}</p>
					
					{#if showRightValue}
						<p class="panel-value animate-reveal" class:correct={isCorrect} class:wrong={isCorrect === false}>
							{formatValue(comparison.right.value, comparison.category)}
						</p>
						<p class="panel-label">{comparison.valueLabel}</p>
					{:else}
						<div class="answer-buttons">
							<button 
								class="answer-btn higher"
								on:click={() => handleAnswer('higher')}
								disabled={isAnimating}
							>
								<span>{higherText}</span>
								<span class="arrow">▲</span>
							</button>
							<button 
								class="answer-btn lower"
								on:click={() => handleAnswer('lower')}
								disabled={isAnimating}
							>
								<span>{lowerText}</span>
								<span class="arrow">▼</span>
							</button>
						</div>
						<p class="compare-text">{thanText} {comparison.left.name}</p>
					{/if}
				</div>
			</div>
			
			<!-- Score Display -->
			<div class="score-overlay score-left">
				<span>{highScoreText}: {highScore}</span>
			</div>
			<div class="score-overlay score-right">
				<span>{scoreLabelText}</span>
			</div>
		</div>
{/if}

<style>
	/* Category Grid (used in GameSetupScreen options slot) */
	.category-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
		gap: 1rem;
	}
	
	.category-card {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.5rem;
		padding: 1.5rem 1rem;
		background: var(--color-surface);
		border: 2px solid transparent;
		border-radius: 1rem;
		cursor: pointer;
		transition: all 0.2s;
	}
	
	.category-card:hover {
		border-color: var(--color-primary);
		transform: translateY(-2px);
	}
	
	.category-card.selected {
		border-color: var(--color-primary);
		background: rgba(99, 102, 241, 0.1);
	}
	
	.category-icon {
		font-size: 2.5rem;
	}
	
	.category-label {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--color-text-light);
	}
	
	.category-desc {
		font-size: 0.75rem;
		color: var(--color-text-muted);
		text-align: center;
	}
	
	/* Score Display (for Game Over) */
	.score-display {
		display: flex;
		gap: 3rem;
		justify-content: center;
	}
	
	.final-score, .high-score {
		display: flex;
		flex-direction: column;
		align-items: center;
	}
	
	.score-label {
		font-size: 1rem;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.1em;
	}
	
	.score-value {
		font-size: 4rem;
		font-weight: 700;
		color: var(--color-primary-light);
		font-family: 'Roboto Mono', monospace;
	}
	
	/* Game Screen - Full height split view */
	.game-screen {
		display: flex;
		min-height: calc(100vh - 4rem);
		width: 100%;
		position: relative;
		transition: transform 0.5s ease-in-out;
		overflow: hidden;
	}
	
	.game-screen.slide-left {
		animation: slideLeftAnim 0.5s ease-in-out;
	}
	
	@keyframes slideLeftAnim {
		0% { transform: translateX(0); }
		50% { transform: translateX(-25%); opacity: 0.5; }
		100% { transform: translateX(0); }
	}
	
	.panel {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		position: relative;
		overflow: hidden;
	}
	
	.panel::before {
		content: '';
		position: absolute;
		inset: 0;
		background: rgba(0, 0, 0, 0.3);
	}
	
	.panel-content {
		position: relative;
		z-index: 1;
		text-align: center;
		padding: 2rem;
		color: white;
	}
	
	.panel-image {
		width: 120px;
		height: auto;
		margin: 0 auto 1.5rem;
		border-radius: 0.5rem;
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
	}
	
	.panel-emoji {
		font-size: 5rem;
		display: block;
		margin-bottom: 1.5rem;
	}
	
	.panel-title {
		font-size: 2rem;
		font-weight: 700;
		margin-bottom: 0.5rem;
		text-shadow: 0 2px 10px rgba(0, 0, 0, 0.5);
	}
	
	@media (min-width: 768px) {
		.panel-title {
			font-size: 2.5rem;
		}
	}
	
	.panel-subtitle {
		font-size: 1rem;
		opacity: 0.8;
		margin-bottom: 0.5rem;
	}
	
	.panel-value {
		font-size: 3rem;
		font-weight: 700;
		color: #FFD700;
		font-family: 'Roboto Mono', monospace;
		text-shadow: 0 2px 10px rgba(0, 0, 0, 0.5);
	}
	
	@media (min-width: 768px) {
		.panel-value {
			font-size: 4rem;
		}
	}
	
	.panel-value.correct {
		color: #10B981;
		animation: pulse 0.5s ease-in-out;
	}
	
	.panel-value.wrong {
		color: #EF4444;
		animation: shake 0.5s ease-in-out;
	}
	
	.panel-label {
		font-size: 0.875rem;
		opacity: 0.7;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		margin-top: 0.25rem;
	}
	
	/* VS Circle */
	.vs-circle {
		position: absolute;
		left: 50%;
		top: 50%;
		transform: translate(-50%, -50%);
		width: 70px;
		height: 70px;
		background: white;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 700;
		font-size: 1.25rem;
		color: #1a1a1a;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
		z-index: 10;
	}
	
	/* Answer Buttons */
	.answer-buttons {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		margin: 1rem 0;
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
		color: white;
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
	
	.compare-text {
		font-size: 0.875rem;
		opacity: 0.7;
		margin-top: 0.5rem;
	}
	
	/* Score Overlays */
	.score-overlay {
		position: absolute;
		bottom: 1.5rem;
		padding: 0.5rem 1rem;
		background: rgba(0, 0, 0, 0.5);
		border-radius: 0.5rem;
		color: white;
		font-weight: 600;
		z-index: 10;
	}
	
	.score-left {
		left: 1.5rem;
	}
	
	.score-right {
		right: 1.5rem;
	}
	
	/* Animations */
	.animate-reveal {
		animation: revealValue 0.5s ease-out;
	}
	
	@keyframes revealValue {
		from {
			opacity: 0;
			transform: scale(0.5);
		}
		to {
			opacity: 1;
			transform: scale(1);
		}
	}
	
	@keyframes pulse {
		0%, 100% { transform: scale(1); }
		50% { transform: scale(1.1); }
	}
	
	@keyframes shake {
		0%, 100% { transform: translateX(0); }
		25% { transform: translateX(-5px); }
		75% { transform: translateX(5px); }
	}
	
	/* Mobile Responsive */
	@media (max-width: 768px) {
		.game-screen {
			flex-direction: column;
		}
		
		.vs-circle {
			width: 50px;
			height: 50px;
			font-size: 1rem;
		}
		
		.panel-image {
			width: 80px;
		}
		
		.panel-emoji {
			font-size: 3.5rem;
		}
		
		.panel-title {
			font-size: 1.5rem;
		}
		
		.panel-value {
			font-size: 2.5rem;
		}
		
		.answer-btn {
			padding: 0.75rem 1.5rem;
			min-width: 150px;
			font-size: 1rem;
		}
		
		.score-overlay {
			font-size: 0.75rem;
			padding: 0.375rem 0.75rem;
		}
	}
</style>
