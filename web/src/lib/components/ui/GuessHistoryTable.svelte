<script>
	import StatusIndicator from './StatusIndicator.svelte';
	import TrendIndicator from './TrendIndicator.svelte';
	
	// Define column configuration
	export let columns = [
		{ key: 'flag', label: 'Flag', type: 'flag' },
		{ key: 'country', label: 'Country', type: 'text' },
		{ key: 'continent', label: 'Continent', type: 'status' },
		{ key: 'population', label: 'Population', type: 'trend' },
		{ key: 'area', label: 'Area', type: 'trend' }
	];
	
	export let guesses = []; // Array of guess objects with column data
	
	// Format number with commas
	function formatNumber(num) {
		if (typeof num === 'number') {
			return num.toLocaleString();
		}
		return num;
	}
</script>

<div class="w-full overflow-x-auto">
	<table class="w-full border-collapse">
		<thead>
			<tr class="border-b-2 border-white/20">
				{#each columns as column}
					<th class="text-left py-3 px-4 text-sm font-semibold text-sandy-light uppercase tracking-wide">
						{column.label}
					</th>
				{/each}
			</tr>
		</thead>
		<tbody>
			{#each guesses as guess, index}
				<tr class="border-b border-white/10 hover:bg-white/5 transition-colors">
					{#each columns as column}
						<td class="py-3 px-4">
							{#if column.type === 'flag'}
								{#if guess[column.key]}
									<div class="flex items-center justify-center w-16 h-10 bg-white/10 rounded overflow-hidden">
										<img 
											src={guess[column.key]} 
											alt={`${guess.country || ''} flag`}
											class="w-full h-full object-cover"
										/>
									</div>
								{/if}
							{:else if column.type === 'text'}
								<span class="text-sandy-light text-sm font-medium">{guess[column.key] || '-'}</span>
							{:else if column.type === 'status'}
								<StatusIndicator 
									status={guess[column.key]?.status === 'match' || guess[column.key] === 'match' ? 'match' : 'no-match'} 
									variant={guess[column.key]?.text ? 'text' : 'square'}
									text={guess[column.key]?.text || guess[column.key]}
								/>
							{:else if column.type === 'trend'}
								<TrendIndicator 
									value={formatNumber(guess[column.key]?.value ?? guess[column.key] ?? '')}
									direction={guess[column.key]?.direction ?? 'none'}
								/>
							{/if}
						</td>
					{/each}
				</tr>
			{/each}
		</tbody>
	</table>
	
	{#if guesses.length === 0}
		<div class="text-center py-12 text-sandy-light">
			<p class="text-sm">No guesses yet</p>
		</div>
	{/if}
</div>

<style>
	/* Mobile responsive: allow horizontal scroll on small screens */
	@media (max-width: 640px) {
		table {
			min-width: 600px;
		}
	}
</style>

