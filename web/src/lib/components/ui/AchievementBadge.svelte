<script>
	export let tier = 'bronze';
	export let icon = '🏆';
	export let label = '';
	export let unlocked = true;
	export let active = false;
	
	$: tierColors = {
		bronze: 'bg-bronze/20 border-bronze text-bronze-light',
		silver: 'bg-slate/20 border-slate text-slate-light',
		gold: 'bg-sunset/20 border-sunset text-sunset'
	};
	
	$: badgeClass = unlocked 
		? (active ? `${tierColors[tier]} shadow-glow opacity-100` : `${tierColors[tier]} opacity-100`)
		: 'bg-slate/10 border-slate/30 text-slate/40 grayscale opacity-60';
</script>

<div class="inline-flex flex-col items-center gap-1 p-4 rounded-full border-2 {badgeClass}" class:animate-bounce-gentle={unlocked && active}>
	<span class="text-4xl">{icon}</span>
	{#if label}
		<span class="text-xs font-semibold uppercase tracking-wide mt-1">{label}</span>
	{/if}
	{#if active && unlocked}
		<span class="text-xs text-sage-light font-medium mt-0.5">Active</span>
	{/if}
	{#if !unlocked}
		<span class="text-xs text-slate/60 font-medium mt-0.5">Locked</span>
	{/if}
</div>

