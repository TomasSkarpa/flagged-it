<script>
	import Timer from './Timer.svelte';
	
	export let time = 0;
	export let maxTime = 60;
	export let currentRound = 1;
	export let totalRounds = 10;
</script>

<div class="flex items-center justify-between gap-6 p-4 rounded-lg backdrop-blur-sm border-2 border-white/20 bg-white/10">
	<div class="flex items-center gap-4">
		<Timer {time} {maxTime} />
		<div class="flex flex-col">
			<p class="text-xs text-sandy uppercase tracking-wide font-semibold">Round</p>
			<p class="text-xl font-bold stat-number text-sandy-light round-counter-hud">{currentRound} / {totalRounds}</p>
		</div>
	</div>
</div>

<style>
	/* Light mode: round counter text should be dark */
	:global(:root.light) .round-counter-hud {
		color: #0F172A !important;
	}
	:global(:root.light) .text-xs.text-sandy {
		color: #64748B !important;
	}
</style>


