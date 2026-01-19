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
	export let columns: number | 'auto' = 'auto';

	const dispatch = createEventDispatcher<{
		select: { value: string };
	}>();

	function handleSelect(value: string) {
		selected = value;
		dispatch('select', { value });
	}
</script>

<div class="difficulty-selector">
	<div class="category-grid" class:auto-columns={columns === 'auto'}>
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
		grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
		gap: 1rem;
	}

	.category-grid.auto-columns {
		grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
	}

</style>
