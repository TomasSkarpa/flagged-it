<script>
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
	class="btn-primary w-full {disabled ? 'opacity-50 cursor-not-allowed' : ''}"
	aria-label="Confirm guess"
>
	Confirm Guess
</button>


