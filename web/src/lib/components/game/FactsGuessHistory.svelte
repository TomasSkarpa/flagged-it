<script lang="ts">
	import type { GuessHistoryEntry } from '$lib/api/factsGame';

	export let guesses: GuessHistoryEntry[] = [];

	function isSkipEntry(entry: GuessHistoryEntry): boolean {
		return entry.guess.toLowerCase().trim() === 'skip';
	}

	function getStatusIcon(entry: GuessHistoryEntry): string {
		if (entry.isCorrect === true) {
			return '✓';
		} else if (isSkipEntry(entry)) {
			return '⏭️';
		} else if (entry.isCorrect === false) {
			return '✗';
		}
		return '';
	}

	function getStatusClass(entry: GuessHistoryEntry): string {
		if (entry.isCorrect === true) {
			return 'text-success';
		} else if (isSkipEntry(entry)) {
			return 'text-text-muted';
		} else if (entry.isCorrect === false) {
			return 'text-error';
		}
		return 'text-text-muted';
	}

	function getRowClass(entry: GuessHistoryEntry): string {
		if (entry.isCorrect === true) {
			return 'bg-success/10 border-success';
		} else if (isSkipEntry(entry)) {
			return 'bg-white/5 border-white/20';
		} else if (entry.isCorrect === false) {
			return 'bg-error/10 border-error';
		}
		return '';
	}
</script>

{#if guesses.length > 0}
	<div class="w-full">
		<h3 class="text-lg font-semibold text-sandy-light mb-4">Previous Guesses</h3>
		<div class="card-game overflow-x-auto">
			<table class="w-full border-collapse">
				<thead>
					<tr class="border-b-2 border-white/20">
						<th class="px-4 py-3 text-left font-semibold">Status</th>
						<th class="px-4 py-3 text-left font-semibold">Flag</th>
						<th class="px-4 py-3 text-left font-semibold">Country</th>
					</tr>
				</thead>
				<tbody>
					{#each guesses.slice().reverse() as entry (entry.guess + entry.fact)}
						<tr class="border-b border-white/10 hover:bg-white/5 {getRowClass(entry)}">
							<td class="px-4 py-3">
								<span class="text-xl {getStatusClass(entry)}">
									{getStatusIcon(entry)}
								</span>
							</td>
							<td class="px-4 py-3">
								{#if entry.country}
									<img 
										src={entry.country.flagUrl} 
										alt={entry.country.name}
										class="w-12 h-8 object-cover rounded"
									/>
								{:else if isSkipEntry(entry)}
									<span class="text-text-muted">—</span>
								{:else}
									<span class="text-text-muted">?</span>
								{/if}
							</td>
							<td class="px-4 py-3 font-semibold">
								{#if entry.country}
									{entry.country.name}
								{:else if isSkipEntry(entry)}
									<span class="text-text-muted">Skip / I don't know</span>
								{:else}
									<span class="text-text-muted italic">{entry.guess}</span>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>
{/if}

<style>
	/* Light mode styles for table */
	:global(:root.light) .card-game table thead tr {
		border-color: rgba(0, 0, 0, 0.2) !important;
	}
	:global(:root.light) .card-game table th {
		color: #0F172A !important;
	}
	:global(:root.light) .card-game table tbody tr {
		border-color: rgba(0, 0, 0, 0.1) !important;
	}
	:global(:root.light) .card-game table tbody tr:hover {
		background-color: rgba(0, 0, 0, 0.03) !important;
	}
	:global(:root.light) .card-game table td {
		color: #1E293B !important;
	}
</style>
