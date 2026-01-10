<script>
	export let padding = 'md';
	
	// Extract class from $$restProps and combine with default classes
	$: defaultClasses = `card-game ${padding === 'sm' ? 'p-4' : padding === 'lg' ? 'p-12' : ''}`;
	$: extraClass = $$restProps.class || '';
	$: allClasses = `${defaultClasses} ${extraClass}`.trim();
</script>

<div class={allClasses}>
	<slot />
</div>

