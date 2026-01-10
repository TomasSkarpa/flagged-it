<script>
	export let rank = 1;
	export let player = '';
	export let score = 0;
	export let total = 10;
	export let isCurrentUser = false;
	
	$: percentage = total > 0 ? Math.round((score / total) * 100) : 0;
	$: rankClass = rank === 1 ? 'text-sunset' : rank === 2 ? 'text-slate-light' : rank === 3 ? 'text-bronze' : 'text-sandy';
</script>

<div class="card-stats {isCurrentUser ? 'border-sunset bg-sunset/10' : ''}">
	<div class="flex items-center gap-4 w-full">
		<div class="flex-shrink-0 w-8 text-center">
			<span class="text-lg font-bold {rankClass}">#{rank}</span>
		</div>
		<div class="flex-1 min-w-0">
			<p class="text-base font-semibold text-sandy-light truncate">{player}</p>
		</div>
		<div class="flex-shrink-0 text-right">
			<p class="text-lg font-bold stat-number text-sandy-light">{score}/{total}</p>
			<p class="text-xs text-slate-light">{percentage}%</p>
		</div>
	</div>
</div>


