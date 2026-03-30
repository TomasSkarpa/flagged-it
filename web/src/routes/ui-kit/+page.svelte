<script lang="ts">
	// @ts-nocheck
	export const params: Record<string, string> = {}; // SvelteKit passes this prop (for external reference only)
	
	// Component imports
	import Button from '$lib/components/ui/Button.svelte';
	import GameCard from '$lib/components/ui/GameCard.svelte';
	import CountryCard from '$lib/components/ui/CountryCard.svelte';
	import StatsPanel from '$lib/components/ui/StatsPanel.svelte';
	import Timer from '$lib/components/ui/Timer.svelte';
	import DistanceIndicator from '$lib/components/ui/DistanceIndicator.svelte';
	import StreakCounter from '$lib/components/ui/StreakCounter.svelte';
	import AchievementBadge from '$lib/components/ui/AchievementBadge.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import HintSystem from '$lib/components/ui/HintSystem.svelte';
	import FlagGrid from '$lib/components/ui/FlagGrid.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import ProgressBar from '$lib/components/ui/ProgressBar.svelte';
	import RegionSelector from '$lib/components/ui/RegionSelector.svelte';
	import ScoreDisplay from '$lib/components/ui/ScoreDisplay.svelte';
	import AnswerButton from '$lib/components/ui/AnswerButton.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import Toast from '$lib/components/ui/Toast.svelte';
	import LeaderboardRow from '$lib/components/ui/LeaderboardRow.svelte';
	import StatusIndicator from '$lib/components/ui/StatusIndicator.svelte';
	import TrendIndicator from '$lib/components/ui/TrendIndicator.svelte';
	import MapWithPin from '$lib/components/ui/MapWithPin.svelte';
	import MapWithPinInfo from '$lib/components/ui/MapWithPinInfo.svelte';
	import MapWithPinPlaceGuess from '$lib/components/ui/MapWithPinPlaceGuess.svelte';
	import MapWithPinCompare from '$lib/components/ui/MapWithPinCompare.svelte';
	import ScoreDistanceFeedback from '$lib/components/ui/ScoreDistanceFeedback.svelte';
	import TimerRoundHUD from '$lib/components/ui/TimerRoundHUD.svelte';
	import Keyboard from '$lib/components/ui/Keyboard.svelte';
	import type { KeyboardLayout, KeyState } from '$lib/components/ui';
	
	// Store and utility imports
	import { locale } from '$lib/stores/locale';
	import { t } from '$lib/translations';
	import { getKeyboardLayoutForLocale } from '$lib/utils/keyboardLayout';
	import { triggerConfetti } from '$lib/utils/confetti';
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';

	// Component state
	let timerValue = 45;
	let maxTimer = 60;
	let distance = 250;
	let streak = 7;
	let toggleHints = false;
	let toggleSound = false;
	let toggleDarkMode = false;
	let toggleModalHints = false;
	let toggleModalSound = false;
	let toggleModalAnimations = false;
	let selectedFlags = [];
	let modalOpen = false;
	let toastShow = false;
	let toastType = 'success';
	let selectedRegion = '';
	let progress = 7;
	let totalRounds = 10;
	let selectedAnswer = null;
	
	// Keyboard demo state
	let keyboardLayout: KeyboardLayout | null = null; // null = auto-detect from locale
	let keyboardKeyStates: Record<string, KeyState> = {};
	let keyboardDisabled = false;
	
	// Confetti debug state
	let confettiLoopEnabled = false;
	let confettiLoopInterval: ReturnType<typeof setInterval> | null = null;
	
	// Reactive values for template
	$: currentLocale = (browser && $locale) ? $locale : 'en';
	$: detectedKeyboardLayout = currentLocale ? getKeyboardLayoutForLocale(currentLocale) : 'english';
	$: uiKitDocumentTitle = t('ui_kit.document_title', undefined, currentLocale);
	$: uiKitMetaDescription = t('ui_kit.meta_description', undefined, currentLocale);
	
	function closeModal() {
		modalOpen = false;
	}
	
	function showToast(type) {
		toastType = type;
		toastShow = true;
		setTimeout(() => { toastShow = false; }, 3000);
	}
	
	function triggerConfettiOnce() {
		triggerConfetti({
			particleCount: 50,
			spread: 70,
			origin: { x: 0.5, y: 0.5 },
			duration: 3000
		});
	}
	
	function toggleConfettiLoop() {
		confettiLoopEnabled = !confettiLoopEnabled;
		if (confettiLoopEnabled) {
			confettiLoopInterval = setInterval(() => {
				triggerConfetti({
					particleCount: 50,
					spread: 70,
					origin: { x: Math.random(), y: Math.random() },
					duration: 3000
				});
			}, 2000); // Trigger every 2 seconds
		} else {
			if (confettiLoopInterval) {
				clearInterval(confettiLoopInterval);
				confettiLoopInterval = null;
			}
		}
	}
	
	onDestroy(() => {
		if (confettiLoopInterval) {
			clearInterval(confettiLoopInterval);
		}
	});
	
	const sampleFlags = [
		{ code: 'US', name: 'United States', flagUrl: '/assets/twemoji_flags_cca2/US.svg' },
		{ code: 'FR', name: 'France', flagUrl: '/assets/twemoji_flags_cca2/FR.svg' },
		{ code: 'JP', name: 'Japan', flagUrl: '/assets/twemoji_flags_cca2/JP.svg' },
		{ code: 'BR', name: 'Brazil', flagUrl: '/assets/twemoji_flags_cca2/BR.svg' },
		{ code: 'DE', name: 'Germany', flagUrl: '/assets/twemoji_flags_cca2/DE.svg' },
		{ code: 'IT', name: 'Italy', flagUrl: '/assets/twemoji_flags_cca2/IT.svg' },
		{ code: 'GB', name: 'United Kingdom', flagUrl: '/assets/twemoji_flags_cca2/GB.svg' },
		{ code: 'CA', name: 'Canada', flagUrl: '/assets/twemoji_flags_cca2/CA.svg' },
		{ code: 'AU', name: 'Australia', flagUrl: '/assets/twemoji_flags_cca2/AU.svg' },
		{ code: 'ES', name: 'Spain', flagUrl: '/assets/twemoji_flags_cca2/ES.svg' },
		{ code: 'MX', name: 'Mexico', flagUrl: '/assets/twemoji_flags_cca2/MX.svg' },
		{ code: 'IN', name: 'India', flagUrl: '/assets/twemoji_flags_cca2/IN.svg' }
	];
</script>

<svelte:head>
	<title>{uiKitDocumentTitle}</title>
	<meta name="description" content={uiKitMetaDescription} />
</svelte:head>

<div class="min-h-screen py-12 px-4 sm:px-6 lg:px-8">
	<div class="max-w-7xl mx-auto">
		<!-- Hero Header -->
		<header class="text-center mb-16 pt-8 pb-4 ">
			<h1 class="text-5xl md:text-6xl font-bold gradient-text mb-4">UI Component Library</h1>
			<p class="text-xl text-text-muted max-w-2xl mx-auto">A comprehensive collection of reusable components for building engaging geography games</p>
		</header>

		<!-- Quick Navigation -->
		<nav class="sticky top-20 z-30 mb-12">
			<div class="card-game p-4 backdrop-blur-xl bg-surface/80 border-white/20">
				<div class="flex flex-wrap gap-2 justify-center">
					<a href="#colors" class="px-4 py-2 rounded-lg bg-surface-light hover:bg-surface text-sm font-medium transition-colors">Colors</a>
					<a href="#buttons" class="px-4 py-2 rounded-lg bg-surface-light hover:bg-surface text-sm font-medium transition-colors">Buttons</a>
					<a href="#cards" class="px-4 py-2 rounded-lg bg-surface-light hover:bg-surface text-sm font-medium transition-colors">Cards</a>
					<a href="#forms" class="px-4 py-2 rounded-lg bg-surface-light hover:bg-surface text-sm font-medium transition-colors">Forms</a>
					<a href="#feedback" class="px-4 py-2 rounded-lg bg-surface-light hover:bg-surface text-sm font-medium transition-colors">Feedback</a>
					<a href="#confetti" class="px-4 py-2 rounded-lg bg-surface-light hover:bg-surface text-sm font-medium transition-colors">Confetti</a>
					<a href="#game" class="px-4 py-2 rounded-lg bg-surface-light hover:bg-surface text-sm font-medium transition-colors">Game Components</a>
				</div>
			</div>
		</nav>

		<div class="space-y-24">
			<!-- Color System -->
			<section id="colors" class="scroll-mt-24">
				<div class="mb-6">
					<h2 class="text-3xl font-bold text-text-light mb-2">Color System</h2>
					<p class="text-text-muted">Our carefully crafted palette designed for modern dark interfaces</p>
				</div>
				<GameCard>
					<div class="space-y-8">
						<!-- Primary Colors -->
						<div>
							<h3 class="text-lg font-semibold text-text-light mb-4">Primary Colors</h3>
							<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
								<div class="text-center">
									<div class="w-full h-32 bg-primary rounded-xl mb-3 shadow-lg shadow-primary/20"></div>
									<p class="font-medium text-text-light">Primary</p>
									<p class="text-xs text-text-muted">#6366F1</p>
								</div>
								<div class="text-center">
									<div class="w-full h-32 bg-secondary rounded-xl mb-3 shadow-lg shadow-secondary/20"></div>
									<p class="font-medium text-text-light">Secondary</p>
									<p class="text-xs text-text-muted">#8B5CF6</p>
								</div>
								<div class="text-center">
									<div class="w-full h-32 bg-accent rounded-xl mb-3 shadow-lg shadow-accent/20"></div>
									<p class="font-medium text-text-light">Accent</p>
									<p class="text-xs text-text-muted">#06B6D4</p>
								</div>
								<div class="text-center">
									<div class="w-full h-32 bg-surface rounded-xl mb-3 border border-white/10"></div>
									<p class="font-medium text-text-light">Surface</p>
									<p class="text-xs text-text-muted">#1E293B</p>
								</div>
							</div>
						</div>

						<!-- Status Colors -->
						<div>
							<h3 class="text-lg font-semibold text-text-light mb-4">Status Colors</h3>
							<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
								<div class="text-center">
									<div class="w-full h-32 bg-success rounded-xl mb-3 shadow-lg shadow-success/20"></div>
									<p class="font-medium text-text-light">Success</p>
									<p class="text-xs text-text-muted">#10B981</p>
								</div>
								<div class="text-center">
									<div class="w-full h-32 bg-error rounded-xl mb-3 shadow-lg shadow-error/20"></div>
									<p class="font-medium text-text-light">Error</p>
									<p class="text-xs text-text-muted">#EF4444</p>
								</div>
								<div class="text-center">
									<div class="w-full h-32 bg-warning rounded-xl mb-3 shadow-lg shadow-warning/20"></div>
									<p class="font-medium text-text-light">Warning</p>
									<p class="text-xs text-text-muted">#F59E0B</p>
								</div>
								<div class="text-center">
									<div class="w-full h-32 bg-info rounded-xl mb-3 shadow-lg shadow-info/20"></div>
									<p class="font-medium text-text-light">Info</p>
									<p class="text-xs text-text-muted">#3B82F6</p>
								</div>
							</div>
						</div>
					</div>
				</GameCard>
			</section>

			<!-- Buttons -->
			<section id="buttons" class="scroll-mt-24">
				<div class="mb-6">
					<h2 class="text-3xl font-bold text-text-light mb-2">Buttons</h2>
					<p class="text-text-muted">Interactive elements for user actions and navigation</p>
				</div>
				<GameCard>
					<div class="space-y-8">
						<div>
							<h3 class="text-lg font-semibold text-text-light mb-4">Variants</h3>
							<div class="flex flex-wrap gap-4 items-center">
								<Button variant="primary">Primary Button</Button>
								<Button variant="secondary">Secondary Button</Button>
								<Button variant="danger">Danger Button</Button>
								<Button variant="icon">⚙️</Button>
							</div>
						</div>
						<div>
							<h3 class="text-lg font-semibold text-text-light mb-4">Sizes</h3>
							<div class="flex flex-wrap gap-4 items-center">
								<Button variant="primary" size="sm">Small</Button>
								<Button variant="primary">Default</Button>
								<Button variant="primary" size="lg">Large</Button>
							</div>
						</div>
						<div>
							<h3 class="text-lg font-semibold text-text-light mb-4">States</h3>
							<div class="flex flex-wrap gap-4 items-center">
								<Button variant="primary">Normal</Button>
								<Button variant="primary" disabled>Disabled</Button>
							</div>
						</div>
					</div>
				</GameCard>
			</section>

			<!-- Confetti Debug -->
			<section id="confetti" class="scroll-mt-24">
				<div class="mb-6">
					<h2 class="text-3xl font-bold text-text-light mb-2">Confetti Debug</h2>
					<p class="text-text-muted">Test confetti effects for debugging purposes</p>
				</div>
				<GameCard>
					<div class="space-y-6">
						<div>
							<h3 class="text-lg font-semibold text-text-light mb-4">Confetti Controls</h3>
							<div class="flex flex-wrap gap-4">
								<Button variant="primary" on:click={triggerConfettiOnce}>
									🎉 Trigger Confetti Once
								</Button>
								<Button 
									variant={confettiLoopEnabled ? "danger" : "secondary"} 
									on:click={toggleConfettiLoop}
								>
									{confettiLoopEnabled ? "⏸️ Stop Loop" : "▶️ Start Loop"}
								</Button>
							</div>
							{#if confettiLoopEnabled}
								<p class="mt-4 text-sm text-text-muted">
									Confetti loop is active - triggering every 2 seconds from random positions
								</p>
							{/if}
						</div>
					</div>
				</GameCard>
			</section>

			<!-- Cards -->
			<section id="cards" class="scroll-mt-24">
				<div class="mb-6">
					<h2 class="text-3xl font-bold text-text-light mb-2">Cards</h2>
					<p class="text-text-muted">Container components for organizing content</p>
				</div>
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
					<GameCard padding="md">
						<h3 class="text-xl font-semibold text-text-light mb-2">Game Card</h3>
						<p class="text-text-muted text-sm">Primary container with frosted glass effect and subtle borders</p>
					</GameCard>
					<CountryCard 
						countryName="France" 
						flagUrl="/assets/twemoji_flags_cca2/FR.svg" 
						stat="Population: 67M" 
						selected={false} 
					/>
					<CountryCard 
						countryName="Japan" 
						flagUrl="/assets/twemoji_flags_cca2/JP.svg" 
						stat="Population: 125M" 
						selected={true} 
					/>
				</div>
			</section>

			<!-- Stats & Metrics -->
			<section class="scroll-mt-24">
				<div class="mb-6">
					<h2 class="text-3xl font-bold text-text-light mb-2">Stats & Metrics</h2>
					<p class="text-text-muted">Display numerical data and progress indicators</p>
				</div>
				<div class="space-y-8">
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Stats Panels</h3>
						<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
							<StatsPanel icon="⭐" label="Score" value="1,250" />
							<StatsPanel icon="🎯" label="Accuracy" value="87%" accent={true} />
							<StatsPanel icon="🏆" label="Rank" value="#42" />
						</div>
					</GameCard>
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Progress Indicators</h3>
						<div class="space-y-6">
							<div>
								<p class="text-sm text-text-muted mb-2">Early Game</p>
								<ProgressBar current={2} total={10} label="Round Progress" color="terracotta" />
							</div>
							<div>
								<p class="text-sm text-text-muted mb-2">Mid Game</p>
								<ProgressBar current={5} total={10} label="Round Progress" color="terracotta" />
							</div>
							<div>
								<p class="text-sm text-text-muted mb-2">Late Game</p>
								<ProgressBar current={8} total={10} label="Round Progress" color="terracotta" />
							</div>
						</div>
					</GameCard>
				</div>
			</section>

			<!-- Forms & Inputs -->
			<section id="forms" class="scroll-mt-24">
				<div class="mb-6">
					<h2 class="text-3xl font-bold text-text-light mb-2">Forms & Inputs</h2>
					<p class="text-text-muted">Interactive form controls and selection components</p>
				</div>
				<div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Toggle Switch</h3>
						<div class="space-y-4">
							<Toggle bind:checked={toggleHints} label="Enable hints" />
							<Toggle bind:checked={toggleSound} label="Sound effects" />
							<Toggle bind:checked={toggleDarkMode} label="Dark mode" />
						</div>
					</GameCard>
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Region Selector</h3>
						<RegionSelector
							regions={['Europe', 'Asia', 'Americas', 'Africa', 'Oceania']}
							bind:selected={selectedRegion}
						/>
					</GameCard>
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Flag Selection</h3>
						<FlagGrid flags={sampleFlags} bind:selected={selectedFlags} />
					</GameCard>
					<GameCard class="col-span-1 lg:col-span-2">
						<h3 class="text-lg font-semibold text-text-light mb-4">Keyboard Component</h3>
						<div class="space-y-6">
							<div class="p-3 bg-surface-light/50 rounded-lg border border-primary/20">
								<p class="text-xs text-text-muted mb-1">Auto-Detection</p>
								<p class="text-sm text-text-light">
									Keyboard layout automatically matches your selected language ({currentLocale ? currentLocale.toUpperCase() : 'EN'}: 
									{detectedKeyboardLayout || 'english'}). 
									You can manually override below.
								</p>
							</div>
							<div>
								<p class="text-sm text-text-muted mb-3">Layout Selection (Manual Override)</p>
								<div class="space-y-4">
									<div>
										<p class="text-xs text-text-muted mb-2">Latin Scripts (QWERTY-based)</p>
										<div class="flex flex-wrap gap-2">
											<Button 
												variant={keyboardLayout === null ? 'primary' : 'secondary'} 
												size="sm"
												on:click={() => { keyboardLayout = null; keyboardKeyStates = {}; }}
											>
												Auto ({detectedKeyboardLayout})
											</Button>
											<Button 
												variant={keyboardLayout === 'english' ? 'primary' : 'secondary'} 
												size="sm"
												on:click={() => { keyboardLayout = 'english'; keyboardKeyStates = {}; }}
											>
												English
											</Button>
											<Button 
												variant={keyboardLayout === 'spanish' ? 'primary' : 'secondary'} 
												size="sm"
												on:click={() => { keyboardLayout = 'spanish'; keyboardKeyStates = {}; }}
											>
												Spanish
											</Button>
											<Button 
												variant={keyboardLayout === 'turkish' ? 'primary' : 'secondary'} 
												size="sm"
												on:click={() => { keyboardLayout = 'turkish'; keyboardKeyStates = {}; }}
											>
												Turkish
											</Button>
											<Button 
												variant={keyboardLayout === 'vietnamese' ? 'primary' : 'secondary'} 
												size="sm"
												on:click={() => { keyboardLayout = 'vietnamese'; keyboardKeyStates = {}; }}
											>
												Vietnamese
											</Button>
											<Button 
												variant={keyboardLayout === 'polish' ? 'primary' : 'secondary'} 
												size="sm"
												on:click={() => { keyboardLayout = 'polish'; keyboardKeyStates = {}; }}
											>
												Polish
											</Button>
											<Button 
												variant={keyboardLayout === 'czech' ? 'primary' : 'secondary'} 
												size="sm"
												on:click={() => { keyboardLayout = 'czech'; keyboardKeyStates = {}; }}
											>
												Czech
											</Button>
											<Button 
												variant={keyboardLayout === 'indonesian' ? 'primary' : 'secondary'} 
												size="sm"
												on:click={() => { keyboardLayout = 'indonesian'; keyboardKeyStates = {}; }}
											>
												Indonesian
											</Button>
										</div>
									</div>
									<div>
										<p class="text-xs text-text-muted mb-2">Cyrillic & Other Scripts</p>
										<div class="flex flex-wrap gap-2">
											<Button 
												variant={keyboardLayout === 'russian' ? 'primary' : 'secondary'} 
												size="sm"
												on:click={() => { keyboardLayout = 'russian'; keyboardKeyStates = {}; }}
											>
												Russian
											</Button>
											<Button 
												variant={keyboardLayout === 'ukrainian' ? 'primary' : 'secondary'} 
												size="sm"
												on:click={() => { keyboardLayout = 'ukrainian'; keyboardKeyStates = {}; }}
											>
												Ukrainian
											</Button>
											<Button 
												variant={keyboardLayout === 'korean' ? 'primary' : 'secondary'} 
												size="sm"
												on:click={() => { keyboardLayout = 'korean'; keyboardKeyStates = {}; }}
											>
												Korean
											</Button>
											<Button 
												variant={keyboardLayout === 'japanese' ? 'primary' : 'secondary'} 
												size="sm"
												on:click={() => { keyboardLayout = 'japanese'; keyboardKeyStates = {}; }}
											>
												Japanese
											</Button>
											<Button 
												variant={keyboardLayout === 'arabic' ? 'primary' : 'secondary'} 
												size="sm"
												on:click={() => { keyboardLayout = 'arabic'; keyboardKeyStates = {}; }}
											>
												Arabic
											</Button>
											<Button 
												variant={keyboardLayout === 'hebrew' ? 'primary' : 'secondary'} 
												size="sm"
												on:click={() => { keyboardLayout = 'hebrew'; keyboardKeyStates = {}; }}
											>
												Hebrew
											</Button>
											<Button 
												variant={keyboardLayout === 'thai' ? 'primary' : 'secondary'} 
												size="sm"
												on:click={() => { keyboardLayout = 'thai'; keyboardKeyStates = {}; }}
											>
												Thai
											</Button>
											<Button 
												variant={keyboardLayout === 'greek' ? 'primary' : 'secondary'} 
												size="sm"
												on:click={() => { keyboardLayout = 'greek'; keyboardKeyStates = {}; }}
											>
												Greek
											</Button>
										</div>
									</div>
								</div>
							</div>
							<div>
								<p class="text-sm text-text-muted mb-3">Interactive Demo</p>
								<Keyboard 
									layout={keyboardLayout}
									keyStates={keyboardKeyStates}
									disabled={keyboardDisabled}
									on:keypress={(e) => {
										const key = e.detail.key;
										// Simulate random correct/incorrect
										const isCorrect = Math.random() > 0.5;
										const newKeyStates = { ...keyboardKeyStates };
										newKeyStates[key] = isCorrect ? 'correct' : 'incorrect';
										keyboardKeyStates = newKeyStates; // Trigger reactivity by reassigning
									}}
								/>
								<div class="flex gap-3 mt-4">
									<Button 
										variant="secondary" 
										size="sm"
										on:click={() => { keyboardKeyStates = {}; }}
									>
										Reset Keys
									</Button>
									<Button 
										variant="secondary" 
										size="sm"
										on:click={() => { keyboardDisabled = !keyboardDisabled; }}
									>
										{keyboardDisabled ? 'Enable' : 'Disable'} Keyboard
									</Button>
								</div>
							</div>
						</div>
					</GameCard>
				</div>
			</section>

			<!-- Feedback Components -->
			<section id="feedback" class="scroll-mt-24">
				<div class="mb-6">
					<h2 class="text-3xl font-bold text-text-light mb-2">Feedback & Notifications</h2>
					<p class="text-text-muted">User feedback, alerts, and status indicators</p>
				</div>
				<div class="space-y-8">
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Toast Notifications</h3>
						<div class="flex flex-wrap gap-3">
							<Button variant="primary" on:click={() => showToast('success')}>Success Toast</Button>
							<Button variant="secondary" on:click={() => showToast('error')}>Error Toast</Button>
							<Button variant="icon" on:click={() => showToast('info')}>Info Toast</Button>
						</div>
					</GameCard>
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Status Indicators</h3>
						<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
							<div>
								<h4 class="text-sm font-medium text-text-muted mb-3">Status</h4>
								<div class="flex items-center gap-6">
									<div class="flex items-center gap-2">
										<StatusIndicator status="match" />
										<span class="text-sm text-text-light">Match</span>
									</div>
									<div class="flex items-center gap-2">
										<StatusIndicator status="no-match" />
										<span class="text-sm text-text-light">No Match</span>
									</div>
								</div>
							</div>
							<div>
								<h4 class="text-sm font-medium text-text-muted mb-3">Trends</h4>
								<div class="flex items-center gap-4 flex-wrap">
									<TrendIndicator value="58,927,633" direction="down" />
									<TrendIndicator value="11,165,299" direction="up" />
								</div>
							</div>
						</div>
					</GameCard>
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Feedback Overlays</h3>
						<p class="text-text-muted mb-4">
							Feedback overlays are handled by the <code class="px-2 py-1 bg-white/5 rounded">GameContainer</code> component's built-in <code class="px-2 py-1 bg-white/5 rounded">FeedbackOverlay</code>.
						</p>
						<p class="text-sm text-text-muted">
							See the Game Components section below for examples of feedback in action.
						</p>
					</GameCard>
				</div>
			</section>

			<!-- Game Components -->
			<section id="game" class="scroll-mt-24">
				<div class="mb-6">
					<h2 class="text-3xl font-bold text-text-light mb-2">Game Components</h2>
					<p class="text-text-muted">Specialized components for game mechanics and interactions</p>
				</div>
				<div class="space-y-12">
					<!-- Game Mode Cards -->
					<div>
						<h3 class="text-xl font-semibold text-text-light mb-4">Game Mode Selector</h3>
						<p class="text-sm text-text-muted mb-4">
							Game mode selection uses <code class="px-2 py-1 bg-white/5 rounded">CategoryCard</code> through the <code class="px-2 py-1 bg-white/5 rounded">DifficultySelector</code> component. See the DifficultySelector example below.
						</p>
					</div>

					<!-- Answer Buttons -->
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Answer Buttons</h3>
						<div class="space-y-6">
							<div>
								<p class="text-sm text-text-muted mb-4">Default State</p>
								<div class="grid grid-cols-2 gap-3">
									<AnswerButton answer="Kosovo" />
									<AnswerButton answer="Albania" />
									<AnswerButton answer="Serbia" />
									<AnswerButton answer="Montenegro" />
								</div>
							</div>
							<div>
								<p class="text-sm text-text-muted mb-4">After Selection (Correct Answer)</p>
								<div class="grid grid-cols-2 gap-3">
									<AnswerButton answer="Kosovo" isCorrect={true} disabled={true} />
									<AnswerButton answer="Albania" isWrong={true} disabled={true} />
									<AnswerButton answer="Serbia" isWrong={true} disabled={true} />
									<AnswerButton answer="Montenegro" isWrong={true} disabled={true} />
								</div>
							</div>
						</div>
					</GameCard>

					<!-- Game HUD Components -->
					<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
						<GameCard>
							<h3 class="text-lg font-semibold text-text-light mb-4">Timer & Round HUD</h3>
							<div class="space-y-4">
								<TimerRoundHUD time={45} maxTime={60} currentRound={3} totalRounds={10} />
								<TimerRoundHUD time={15} maxTime={60} currentRound={7} totalRounds={10} />
							</div>
						</GameCard>
						<GameCard>
							<h3 class="text-lg font-semibold text-text-light mb-4">Score & Distance Feedback</h3>
							<div class="space-y-4">
								<ScoreDistanceFeedback score={7} total={10} distance={250} showPercentage={true} />
								<ScoreDistanceFeedback score={5} total={10} distance={1200} showPercentage={true} />
							</div>
						</GameCard>
					</div>

					<!-- Map Components -->
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Map with Pin - Info Display</h3>
						<p class="text-sm text-text-muted mb-3">Display a pin at a specific location (read-only)</p>
						<MapWithPinInfo 
							lat={48.8566} 
							lng={2.3522} 
							zoom={10}
							title="Paris, France"
							height="300px" 
						/>
					</GameCard>
					
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Map with Pin - Place Guess</h3>
						<p class="text-sm text-text-muted mb-3">Click on map and confirm your guess</p>
						<MapWithPinPlaceGuess 
							lat={20} 
							lng={0} 
							zoom={2}
							question="Where is Tokyo, Japan?"
							confirmLabel="Confirm Location"
							height="300px"
							on:guess={(e) => { console.log('Guess:', e.detail); }}
						/>
					</GameCard>
					
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Map with Pin - Compare Distance</h3>
						<p class="text-sm text-text-muted mb-3">Click to guess, then see distance and line to correct location</p>
						<MapWithPinCompare 
							correctLat={35.6762} 
							correctLng={139.6503}
							zoom={4}
							showCorrectAfterClick={true}
							height="300px"
							on:guess={(e) => { console.log('Guess with distance:', e.detail); }}
						/>
					</GameCard>

					<!-- Hint System -->
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Hint System</h3>
						<HintSystem
							hintLevel={3}
							continent="Europe"
							region="Western Europe"
							proximity="Within 500km"
							cost={10}
						/>
					</GameCard>

					<!-- Score Displays -->
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Score Displays</h3>
						<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
							<ScoreDisplay score={8} total={10} showPercentage={true} showProgress={true} />
							<ScoreDisplay score={5} total={10} showPercentage={true} showProgress={false} />
							<ScoreDisplay score={10} total={10} showPercentage={false} showProgress={true} />
						</div>
					</GameCard>

					<!-- Leaderboard -->
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Leaderboard</h3>
						<div class="space-y-2">
							<LeaderboardRow rank={1} player="Explorer123" score={10} total={10} />
							<LeaderboardRow rank={2} player="GeoMaster" score={9} total={10} />
							<LeaderboardRow rank={3} player="WorldWanderer" score={8} total={10} />
							<LeaderboardRow rank={4} player="You" score={7} total={10} isCurrentUser={true} />
							<LeaderboardRow rank={5} player="MapLover" score={6} total={10} />
						</div>
					</GameCard>

					<!-- Worldle Game Mock -->
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Worldle Game - Guess History Mock</h3>
						<p class="text-sm text-text-muted mb-4">Mock of the Worldle game guess history table for styling reference</p>
						<div class="card-game overflow-x-auto">
							<h2 class="text-2xl font-bold mb-4">Guess History:</h2>
							<table class="w-full border-collapse">
								<thead>
									<tr class="border-b-2 border-white/20">
										<th class="px-4 py-3 text-left font-semibold">Flag</th>
										<th class="px-4 py-3 text-left font-semibold">Country</th>
										<th class="px-4 py-3 text-left font-semibold">Continent</th>
										<th class="px-4 py-3 text-left font-semibold">Population</th>
										<th class="px-4 py-3 text-left font-semibold">Area</th>
									</tr>
								</thead>
								<tbody>
									<tr class="border-b border-white/10 hover:bg-white/5">
										<td class="px-4 py-3">
											<img 
												src="/assets/twemoji_flags_cca2/AU.svg" 
												alt="Australia flag"
												class="w-12 h-8 object-cover rounded"
											/>
										</td>
										<td class="px-4 py-3 font-semibold">Australia</td>
										<td class="px-4 py-3">
											<span class="px-3 py-1 rounded border-2 bg-error/30 text-error border-error">
												Oceania
											</span>
										</td>
										<td class="px-4 py-3">
											<span class="px-3 py-1 rounded border-2 flex items-center gap-2 bg-yellow-500/30 text-yellow-300 border-yellow-500">
												<span>▲</span>
												<span>27.5M</span>
											</span>
										</td>
										<td class="px-4 py-3">
											<span class="px-3 py-1 rounded border-2 flex items-center gap-2 bg-error/30 text-error border-error">
												<span>▼</span>
												<span>7.7M km²</span>
											</span>
										</td>
									</tr>
									<tr class="border-b border-white/10 hover:bg-white/5">
										<td class="px-4 py-3">
											<img 
												src="/assets/twemoji_flags_cca2/JP.svg" 
												alt="Japan flag"
												class="w-12 h-8 object-cover rounded"
											/>
										</td>
										<td class="px-4 py-3 font-semibold">Japan</td>
										<td class="px-4 py-3">
											<span class="px-3 py-1 rounded border-2 bg-error/30 text-error border-error">
												Asia
											</span>
										</td>
										<td class="px-4 py-3">
											<span class="px-3 py-1 rounded border-2 flex items-center gap-2 bg-error/30 text-error border-error">
												<span>▼</span>
												<span>123.2M</span>
											</span>
										</td>
										<td class="px-4 py-3">
											<span class="px-3 py-1 rounded border-2 flex items-center gap-2 bg-error/30 text-error border-error">
												<span>▲</span>
												<span>377.9K km²</span>
											</span>
										</td>
									</tr>
									<tr class="border-b border-white/10 hover:bg-white/5">
										<td class="px-4 py-3">
											<img 
												src="/assets/twemoji_flags_cca2/KE.svg" 
												alt="Kenya flag"
												class="w-12 h-8 object-cover rounded"
											/>
										</td>
										<td class="px-4 py-3 font-semibold">Kenya</td>
										<td class="px-4 py-3">
											<span class="px-3 py-1 rounded border-2 bg-error/30 text-error border-error">
												Africa
											</span>
										</td>
										<td class="px-4 py-3">
											<span class="px-3 py-1 rounded border-2 flex items-center gap-2 bg-error/30 text-error border-error">
												<span>▼</span>
												<span>53.3M</span>
											</span>
										</td>
										<td class="px-4 py-3">
											<span class="px-3 py-1 rounded border-2 flex items-center gap-2 bg-error/30 text-error border-error">
												<span>▲</span>
												<span>580.4K km²</span>
											</span>
										</td>
									</tr>
									<tr class="border-b border-white/10 hover:bg-white/5">
										<td class="px-4 py-3">
											<img 
												src="/assets/twemoji_flags_cca2/FR.svg" 
												alt="France flag"
												class="w-12 h-8 object-cover rounded"
											/>
										</td>
										<td class="px-4 py-3 font-semibold">France</td>
										<td class="px-4 py-3">
											<span class="px-3 py-1 rounded border-2 bg-error/30 text-error border-error">
												Europe
											</span>
										</td>
										<td class="px-4 py-3">
											<span class="px-3 py-1 rounded border-2 flex items-center gap-2 bg-error/30 text-error border-error">
												<span>▼</span>
												<span>66.4M</span>
											</span>
										</td>
										<td class="px-4 py-3">
											<span class="px-3 py-1 rounded border-2 flex items-center gap-2 bg-error/30 text-error border-error">
												<span>▲</span>
												<span>543.9K km²</span>
											</span>
										</td>
									</tr>
								</tbody>
							</table>
						</div>
					</GameCard>
				</div>
			</section>

			<!-- Special Components -->
			<section class="scroll-mt-24">
				<div class="mb-6">
					<h2 class="text-3xl font-bold text-text-light mb-2">Special Components</h2>
					<p class="text-text-muted">Timers, counters, and achievement displays</p>
				</div>
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
					<GameCard>
						<h3 class="text-base font-semibold text-text-light mb-4 text-center">Timer</h3>
						<div class="flex justify-center">
							<Timer time={timerValue} maxTime={maxTimer} />
						</div>
					</GameCard>
					<GameCard>
						<h3 class="text-base font-semibold text-text-light mb-4 text-center">Distance</h3>
						<div class="flex justify-center">
							<DistanceIndicator distance={distance} />
						</div>
					</GameCard>
					<GameCard>
						<h3 class="text-base font-semibold text-text-light mb-4 text-center">Streak</h3>
						<div class="flex justify-center">
							<StreakCounter streak={streak} showGlow={true} />
						</div>
					</GameCard>
					<GameCard>
						<h3 class="text-base font-semibold text-text-light mb-4 text-center">Achievements</h3>
						<div class="flex flex-wrap gap-3 justify-center">
							<AchievementBadge tier="bronze" icon="🥉" label="Bronze" unlocked={true} active={false} />
							<AchievementBadge tier="gold" icon="🥇" label="Gold" unlocked={true} active={true} />
						</div>
					</GameCard>
				</div>
			</section>

			<!-- Loading & Empty States -->
			<section class="scroll-mt-24">
				<div class="mb-6">
					<h2 class="text-3xl font-bold text-text-light mb-2">Loading & Empty States</h2>
					<p class="text-text-muted">States for loading content and empty data</p>
				</div>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<GameCard>
						<h3 class="text-lg font-semibold text-text-light mb-4">Loading Spinners</h3>
						<div class="flex gap-8 justify-center items-center">
							<LoadingSpinner size="sm" />
							<LoadingSpinner size="md" />
							<LoadingSpinner size="lg" />
						</div>
					</GameCard>
					<GameCard>
						<EmptyState
							icon="🧭"
							title="Start Your Journey"
							message="Begin exploring countries and test your geography knowledge!"
							actionLabel="Start Game"
						/>
					</GameCard>
				</div>
			</section>

			<!-- Modal & Overlays -->
			<section class="scroll-mt-24">
				<div class="mb-6">
					<h2 class="text-3xl font-bold text-text-light mb-2">Modals & Overlays</h2>
					<p class="text-text-muted">Dialog components for important interactions</p>
				</div>
				<GameCard>
					<Button variant="primary" on:click={() => { modalOpen = true; }}>
						Open Modal Dialog
					</Button>
				</GameCard>
			</section>
		</div>
	</div>
</div>

<!-- Modal -->
<Modal title="Game Settings" open={modalOpen} onClose={() => closeModal()}>
	<div class="space-y-4">
		<p class="text-text-light">Configure your game preferences here.</p>
		<div class="space-y-3">
			<Toggle bind:checked={toggleModalHints} label="Enable hints" />
			<Toggle bind:checked={toggleModalSound} label="Sound effects" />
			<Toggle bind:checked={toggleModalAnimations} label="Animations" />
		</div>
		<div class="flex gap-3 mt-6">
			<Button variant="primary" on:click={() => modalOpen = false}>Save</Button>
			<Button variant="secondary" on:click={() => modalOpen = false}>Cancel</Button>
		</div>
	</div>
</Modal>

<!-- Toast -->
<Toast message={toastType === 'success' ? 'Operation successful!' : toastType === 'error' ? 'Something went wrong!' : 'Here is some information'} type={toastType} show={toastShow} />
