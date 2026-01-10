<script lang="ts">
	import type { GuessHistoryEntry } from '$lib/api/factsGame';

	export let guesses: GuessHistoryEntry[] = [];

	function formatFact(fact: string): string {
		// Remove "Fact X: " prefix if present (backend includes it)
		const cleanedFact = fact.replace(/^Fact \d+:\s*/, '');
		// Convert markdown-style bold **text** to HTML <strong>text</strong>
		return cleanedFact.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
	}

	function getGuessDisplayText(entry: GuessHistoryEntry): string {
		// Remove checkmark if present
		const cleaned = entry.guess.replace('✅', '').replace('✓', '').trim();
		// Handle skip entries
		if (cleaned.toLowerCase() === 'skip') {
			return '⏭️ Skip / I don\'t know';
		}
		return cleaned;
	}
	
	function isSkipEntry(entry: GuessHistoryEntry): boolean {
		return entry.guess.toLowerCase().trim() === 'skip';
	}
</script>

{#if guesses.length > 0}
	<div class="w-full">
		<h3 class="text-lg font-semibold text-sandy-light mb-4">Previous Guesses</h3>
		<div class="space-y-2">
			{#each guesses.slice().reverse() as entry, index}
				{@const originalIndex = guesses.length - index - 1}
				{@const isSkip = isSkipEntry(entry)}
				<div class="card-game {entry.isCorrect === true ? 'bg-success/10 border-success' : entry.isCorrect === false && !isSkip ? 'bg-error/10 border-error' : isSkip ? 'bg-white/5 border-white/20' : ''}">
					<div class="flex items-center gap-3">
						<div class="flex-shrink-0 w-8 text-center">
							{#if entry.isCorrect === true}
								<span class="text-success text-xl">✓</span>
							{:else if isSkip}
								<span class="text-text-muted text-xl">⏭️</span>
							{:else if entry.isCorrect === false}
								<span class="text-error text-xl">✗</span>
							{:else}
								<span class="text-text-muted text-sm">#{originalIndex + 1}</span>
							{/if}
						</div>
						<div class="flex-1 min-w-0">
							<p class="font-medium {entry.isCorrect === true ? 'text-success' : entry.isCorrect === false && !isSkip ? 'text-error' : isSkip ? 'text-text-muted' : 'text-sandy-light'}">
								{getGuessDisplayText(entry)}
							</p>
							{#if entry.fact}
								<p class="text-sm text-text-muted mt-1 line-clamp-2" innerHTML={formatFact(entry.fact)}></p>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>
	</div>
{/if}
