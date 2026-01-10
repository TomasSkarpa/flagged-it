<script>
	export let time = 0; // in seconds
	export let maxTime = 60;
	
	$: percentage = (time / maxTime) * 100;
	$: minutes = Math.floor(time / 60);
	$: seconds = time % 60;
	$: displayTime = `${minutes}:${seconds.toString().padStart(2, '0')}`;
	
	$: colorClass = percentage > 50 ? 'text-sage' : percentage > 25 ? 'text-sunset' : 'text-pin-red';
	$: ringColor = percentage > 50 ? 'stroke-sage' : percentage > 25 ? 'stroke-sunset' : 'stroke-pin-red';
</script>

<div class="relative inline-flex items-center justify-center" role="timer" aria-label="Time remaining: {displayTime}">
	<!-- Circular progress ring -->
	<svg class="transform -rotate-90 w-24 h-24" aria-hidden="true">
		<circle
			cx="48"
			cy="48"
			r="42"
			fill="none"
			stroke="currentColor"
			stroke-width="6"
			class="text-white/15"
		/>
		<circle
			cx="48"
			cy="48"
			r="42"
			fill="none"
			stroke="currentColor"
			stroke-width="6"
			stroke-dasharray={`${2 * Math.PI * 42}`}
			stroke-dashoffset={`${2 * Math.PI * 42 * (1 - percentage / 100)}`}
			stroke-linecap="round"
			class={ringColor}
			style="transition: stroke-dashoffset 0.3s ease, stroke 0.3s ease; filter: drop-shadow(0 0 4px currentColor);"
		/>
	</svg>
	<!-- Time display -->
	<div class="absolute inset-0 flex items-center justify-center">
		<div class="text-center">
			<span class="text-2xl font-bold stat-number {colorClass}">{displayTime}</span>
		</div>
	</div>
</div>

