<script>
	export let status = 'match'; // 'match' (green) or 'no-match' (red)
	export let size = 'md'; // 'sm', 'md', 'lg'
	export let text = undefined; // Optional text to display with colored background
	export let variant = 'square'; // 'square' (just colored square) or 'text' (text with colored background)
	
	$: bgClass = status === 'match' ? 'bg-sage' : 'bg-pin-red';
	$: sizeClass = size === 'sm' ? 'w-3 h-3' : size === 'lg' ? 'w-5 h-5' : 'w-4 h-4';
	
	$: textClasses = variant === 'text' 
		? `px-2 py-1 rounded text-xs font-medium text-white ${bgClass}` 
		: '';
</script>

{#if variant === 'text' && text}
	<div class="{textClasses}" aria-label={status === 'match' ? 'Match' : 'No match'}>
		{text}
	</div>
{:else}
	<div class="{sizeClass} {bgClass} rounded-sm" aria-label={status === 'match' ? 'Match' : 'No match'}></div>
{/if}

