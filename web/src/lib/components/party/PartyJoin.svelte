<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';

	export let code: string;
	export let name = '';
	export let isLoading = false;
	export let error: string | null = null;

	const dispatch = createEventDispatcher<{ join: { code: string; name: string }; back: void }>();

	$: L = $locale;
</script>

<div class="max-w-md mx-auto text-center px-4">
	<h1 class="text-4xl md:text-5xl font-bold text-sandy-light mb-8">
		<span class="mr-2 emoji-blue">👥</span>
		<span class="gradient-text">{t('party.join.heading', { Code: code }, L)}</span>
	</h1>
	<form class="card-game text-left" on:submit|preventDefault={() => dispatch('join', { code, name: name.trim() })}>
		<label class="block text-sm font-semibold text-sandy mb-2" for="party-join-name">{t('party.name.label', undefined, L)}</label>
		<input
			id="party-join-name"
			type="text"
			bind:value={name}
			maxlength="16"
			placeholder={t('party.name.placeholder', undefined, L)}
			autocomplete="nickname"
			class="party-input w-full mb-6"
		/>
		<button type="submit" class="btn-primary w-full px-8 py-4 text-xl font-bold" disabled={isLoading}>
			{t('party.join.button', undefined, L)}
		</button>
	</form>
	{#if error}
		<div class="mt-6 p-4 bg-error/20 border border-error rounded-lg">
			<p class="text-error font-semibold">{error}</p>
		</div>
	{/if}
	<button class="block mx-auto mt-8 text-text-muted underline text-sm" on:click={() => dispatch('back')}>
		{t('party.new_game', undefined, L)}
	</button>
</div>
