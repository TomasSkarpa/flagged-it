<script lang="ts">
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';

	// @ts-nocheck
	export let status = 500;
	export let error = null;
	// svelte-ignore unused-export-let
	export const params = {}; // SvelteKit passes this prop
	
	// Type assertion for error to access message property
	$: currentLocale = $locale;
	$: errorTitle = t('error.title', undefined, currentLocale);
	$: defaultErrorMessage = t('error.message', undefined, currentLocale);
	$: goHomeText = t('error.go_home', undefined, currentLocale);
	$: errorMessage = error && typeof error === 'object' && 'message' in error ? error['message'] : defaultErrorMessage;
</script>

<svelte:head>
	<title>{status} - {errorTitle}</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center px-4">
	<div class="text-center">
		<h1 class="text-6xl font-bold text-text-light mb-4">{status}</h1>
		<p class="text-xl text-text-muted mb-8">{errorMessage}</p>
		<a href="/" class="btn-primary">{goHomeText}</a>
	</div>
</div>
