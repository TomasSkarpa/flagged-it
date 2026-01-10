<script>
	export let answer = '';
	export let isCorrect = false;
	export let isWrong = false;
	export let disabled = false;
	export let onClick = null;
	
	function handleClick() {
		if (onClick && typeof onClick === 'function') {
			onClick();
		}
	}
</script>

<button
	on:click={handleClick}
	disabled={disabled}
	class="w-full px-6 py-4 rounded-lg border-2 font-semibold text-left transition-all duration-200 {isCorrect ? 'bg-sage border-sage text-white' : isWrong ? 'bg-pin-red border-pin-red text-white' : 'bg-white/10 border-white/20 text-sandy-light hover:border-sky hover:bg-sky/10'} disabled:opacity-50 disabled:cursor-not-allowed"
>
	{answer}
</button>


