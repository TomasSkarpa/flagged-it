<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Country } from '$lib/types';
	import { getCountryNameForLocale } from '$lib/utils/countryNames';
	import { locale } from '$lib/stores/locale';

	export let options: Country[] | string[];
	export let selectedAnswer: string | null = null;
	export let correctAnswer: string | null = null;
	export let showFeedback: boolean = false;
	export let isCorrect: boolean = false;
	export let disabled: boolean = false;
	export let columns: 1 | 2 = 2;

	const dispatch = createEventDispatcher<{
		select: { country?: Country; answer?: string };
	}>();

	// Make component reactive to locale changes
	$: currentLocale = $locale;

	function handleSelect(option: Country | string) {
		if (!showFeedback && !disabled) {
			if (typeof option === 'string') {
				dispatch('select', { answer: option });
			} else {
				dispatch('select', { country: option });
			}
		}
	}

	function getButtonClass(option: Country | string): string {
		const optionValue = typeof option === 'string' ? option : option.cca2;
		const optionName = typeof option === 'string' ? option : getOptionLabel(option);

		const isSelected = selectedAnswer === optionValue;
		const isCorrectOption = showFeedback && optionName === correctAnswer;
		const isWrongOption = showFeedback && isSelected && !isCorrect;

		if (isCorrectOption) {
			return 'bg-success border-success text-white shadow-glow';
		}
		if (isWrongOption) {
			return 'bg-error border-error text-white animate-shake';
		}
		if (isSelected) {
			return 'bg-primary/30 border-primary text-sandy-light';
		}
		return 'bg-white/10 border-white/20 text-sandy-light hover:border-accent hover:bg-accent/10';
	}

	function getOptionKey(option: Country | string, index: number): string {
		if (typeof option === 'string') {
			return option;
		}
		return option.cca2 || `option-${index}`;
	}

	function getOptionLabel(option: Country | string): string {
		if (typeof option === 'string') {
			return option;
		}
		// Use translated country name (backend already translated based on locale)
		// Reference currentLocale to make this reactive to locale changes
		return getCountryNameForLocale(option);
	}
</script>

<div class="grid gap-4 {columns === 1 ? 'grid-cols-1' : 'grid-cols-1 md:grid-cols-2'}">
	{#each options as option, index (getOptionKey(option, index))}
		<button
			on:click={() => handleSelect(option)}
			disabled={showFeedback || disabled}
			class="w-full px-6 py-4 rounded-lg border-2 font-semibold text-left transition-all duration-200 
				{getButtonClass(option)} 
				disabled:opacity-50 disabled:cursor-not-allowed"
		>
			{getOptionLabel(option)}
		</button>
	{/each}
</div>
