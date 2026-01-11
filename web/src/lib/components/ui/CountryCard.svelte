<script>
	export let selected = false;
	export let flagUrl = '';
	export let countryName = '';
	export let stat = '';
	
	$: hasFlag = flagUrl && flagUrl !== '';
</script>

<div class="card-country card-country-library {selected ? 'selected' : ''}">
	{#if hasFlag}
		<div class="flag-container">
			<img src={flagUrl} alt={`${countryName} flag`} class="flag-image" />
		</div>
	{/if}
	<div class="card-content">
		<h3 class="card-title">{countryName}</h3>
		{#if stat}
			<p class="card-stat">{stat}</p>
		{/if}
		<slot />
	</div>
</div>

<style>
	.card-country-library {
		display: flex;
		flex-direction: column;
		height: 100%;
		min-height: 200px;
	}
	
	.flag-container {
		flex-shrink: 0;
		height: 96px;
		background: rgba(255, 255, 255, 0.1);
		border-radius: 0.5rem;
		padding: 0.5rem;
		margin-bottom: 0.75rem;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	:global(:root.light) .flag-container {
		background: rgba(0, 0, 0, 0.06);
	}
	
	.flag-image {
		width: 100%;
		height: 100%;
		object-fit: contain;
		border-radius: 0.25rem;
	}
	
	.card-content {
		display: flex;
		flex-direction: column;
		flex: 1;
		min-height: 0;
	}
	
	.card-title {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--color-text-light);
		margin-bottom: 0.25rem;
		flex-shrink: 0;
		line-height: 1.3;
	}
	
	.card-stat {
		font-size: 0.875rem;
		color: var(--color-text-muted);
		font-family: 'Roboto Mono', monospace;
		flex-shrink: 0;
		margin-top: auto;
		padding-top: 0.5rem;
	}
</style>

