<script>
	export let padding = 'md';
</script>

<div class="card-game {padding === 'sm' ? 'p-4' : padding === 'lg' ? 'p-12' : ''}">
	<slot />
</div>

