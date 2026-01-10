<script>
	export let size = 'md';
	
	$: sizeClass = size === 'sm' ? 'w-8 h-8' : size === 'lg' ? 'w-16 h-16' : 'w-12 h-12';
</script>

<div class="flex items-center justify-center">
	<div class="{sizeClass} border-4 border-white/20 border-t-terracotta rounded-full animate-spin"></div>
</div>

