<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export let selected: boolean = false;
	export let icon: string | null = null;
	export let label: string = '';
	export let description: string = '';

	const dispatch = createEventDispatcher();
</script>

<button
	class="category-card"
	class:selected
	type="button"
	on:click
>
	{#if icon}
		<span class="category-icon">{icon}</span>
	{/if}
	<span class="category-label">{label}</span>
	{#if description}
		<span class="category-desc">{description}</span>
	{/if}
</button>

<style>
	.category-card {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.5rem;
		padding: 1.5rem 1rem;
		background: var(--color-surface);
		border: 2px solid transparent;
		border-radius: 1rem;
		cursor: pointer;
		transition: all 0.2s;
		width: 100%;
	}

	.category-card:hover {
		border-color: var(--color-primary);
		transform: translateY(-2px);
	}

	.category-card.selected {
		border-color: var(--color-primary);
		background: rgba(99, 102, 241, 0.1);
	}

	.category-icon {
		font-size: 2.5rem;
	}

	.category-label {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--color-text-light);
	}

	.category-desc {
		font-size: 0.75rem;
		color: var(--color-text-muted);
		text-align: center;
	}

	/* Light Mode Styles */
	:global(:root.light) .category-card {
		background: rgba(255, 255, 255, 0.9);
		border-color: rgba(15, 23, 42, 0.2);
	}

	:global(:root.light) .category-card:hover {
		border-color: var(--color-primary);
		background: rgba(255, 255, 255, 1);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	:global(:root.light) .category-card.selected {
		border-color: var(--color-primary);
		background: rgba(99, 102, 241, 0.15);
	}

	:global(:root.light) .category-label {
		color: #0F172A;
	}

	:global(:root.light) .category-desc {
		color: #64748B;
	}
</style>
