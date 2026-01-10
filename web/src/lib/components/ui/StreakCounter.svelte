<script>
	export let streak = 0;
	export let showGlow = false;
	
	$: isMilestone = streak > 0 && streak % 5 === 0;
</script>

<div class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-white/10 backdrop-blur-sm border border-white/20" class:text-glow={showGlow || isMilestone}>
	<span class="text-2xl">🔥</span>
	<div>
		<span class="text-2xl font-bold stat-number text-sunset-light">{streak}</span>
		<span class="text-xs text-sandy ml-1 uppercase tracking-wide font-medium">streak</span>
	</div>
</div>

