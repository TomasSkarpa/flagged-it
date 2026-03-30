<script context="module" lang="ts">
	export type DifficultyOption = {
		value: string;
		label: string;
		description: string;
	};
</script>

<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import CategoryCard from '$lib/components/ui/CategoryCard.svelte';

	type DifficultyOption = {
		value: string;
		label: string;
		description: string;
		icon?: string;
	};

	export let options: DifficultyOption[] = [];
	export let selected: string = '';

	const dispatch = createEventDispatcher<{
		select: { value: string };
	}>();

	function handleSelect(value: string) {
		selected = value;
		dispatch('select', { value });
	}
</script>

<div class="difficulty-selector">
	<div class="category-grid">
		{#each options as option}
			<CategoryCard
				selected={selected === option.value}
				icon={option.icon || null}
				label={option.label}
				description={option.description}
				on:click={() => handleSelect(option.value)}
			/>
		{/each}
	</div>
</div>

<style>
	.difficulty-selector {
		width: 100%;
	}

	.category-grid {
		display: grid;
		gap: 1rem;
		/* Default: 3 columns (5+ items, or if :has() is unsupported) */
		grid-template-columns: repeat(3, minmax(0, 1fr));
	}

	/* Exactly 4 → 2×2 */
	.category-grid:has(> :nth-child(4):last-child) {
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}

	/* Exactly 3 → one row */
	.category-grid:has(> :nth-child(3):last-child) {
		grid-template-columns: repeat(3, minmax(0, 1fr));
	}

	/* Exactly 2 */
	.category-grid:has(> :nth-child(2):last-child) {
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}

	/* Single card */
	.category-grid:has(> :only-child) {
		grid-template-columns: minmax(0, 1fr);
	}

	@media (max-width: 480px) {
		.difficulty-selector .category-grid {
			grid-template-columns: minmax(0, 1fr);
		}
	}

</style>
