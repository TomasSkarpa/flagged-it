<script lang="ts">
	import { page } from '$app/stores';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	import {
		get404IllustrationSrc,
		pickRandom404Variant,
		type Error404Variant
	} from '$lib/utils/error-404-variant';

	// SvelteKit puts the real HTTP status on the page store; `export let status` is not set by the
	// framework for root +error.svelte, so defaulting it to 500 incorrectly showed the 500 UI for 404s.

	$: currentLocale = $locale;
	$: status = $page.status;
	$: kitError = $page.error;
	$: isServerMeltdown = status === 500;
	$: isNotFound = status === 404;

	let variant404: Error404Variant | null = null;
	let variant404PathKey = '';
	$: {
		const nextKey = status === 404 ? $page.url.pathname : '';
		if (nextKey !== variant404PathKey) {
			variant404PathKey = nextKey;
			variant404 = nextKey ? pickRandom404Variant() : null;
		}
	}
	$: img404 = variant404 ? get404IllustrationSrc(variant404) : '';
	$: line1Key = variant404 != null ? `error.404.${variant404}.line1` : '';
	$: line2Key = variant404 != null ? `error.404.${variant404}.line2` : '';
	$: line1404 =
		variant404 != null && line1Key
			? t(line1Key, undefined, currentLocale)
			: '';
	$: line2404Raw =
		variant404 != null && line2Key ? t(line2Key, undefined, currentLocale) : '';
	$: line2404 =
		variant404 != null && line2Key && line2404Raw !== line2Key ? line2404Raw : '';
	$: button404 =
		variant404 != null ? t(`error.404.${variant404}.button`, undefined, currentLocale) : '';
	$: alt404 =
		variant404 != null ? t(`error.404.${variant404}.image_alt`, undefined, currentLocale) : '';
	$: errorTitle = t('error.title', undefined, currentLocale);
	$: defaultErrorMessage = t('error.message', undefined, currentLocale);
	$: goHomeText = t('error.go_home', undefined, currentLocale);
	$: errorMessage =
		kitError && typeof kitError === 'object' && 'message' in kitError
			? kitError['message']
			: defaultErrorMessage;
	$: metaTitle = isServerMeltdown
		? t('error.500.meta_title', undefined, currentLocale)
		: isNotFound && variant404 != null
			? t(`error.404.${variant404}.meta_title`, undefined, currentLocale)
			: `${status} - ${errorTitle}`;
	$: heading500 = t('error.500.heading', undefined, currentLocale);
	$: message500 = t('error.500.message', undefined, currentLocale);
</script>

<svelte:head>
	<title>{metaTitle}</title>
</svelte:head>

<!-- Match +layout.svelte main: min-height calc(100vh - 4rem); 404 is full-bleed -->
<div
	class="min-h-[calc(100vh-4rem)] w-full flex flex-col {isNotFound && variant404 != null
		? 'items-stretch px-0 py-0'
		: 'items-center justify-center px-4 py-6 sm:py-10'}"
>
	{#if isServerMeltdown}
		<div
			class="max-w-md w-full mx-auto rounded-2xl border border-white/10 bg-surface/90 backdrop-blur-md px-6 sm:px-10 py-9 sm:py-11 text-center shadow-2xl shadow-black/20 ring-1 ring-inset ring-white/5"
		>
			<div class="flex justify-center mb-5 sm:mb-6" aria-hidden="true">
				<img
					src="/assets/errors/surrender-flag.svg"
					alt=""
					width="100"
					height="110"
					class="opacity-95 drop-shadow-[0_8px_24px_rgba(0,0,0,0.35)] motion-safe:animate-[gentle-wave_2.8s_ease-in-out_infinite]"
				/>
			</div>
			<p
				class="inline-block text-xs font-semibold uppercase tracking-[0.2em] text-accent mb-3 px-2 py-1 rounded-full bg-accent/10 border border-accent/20"
			>
				HTTP 500
			</p>
			<h1 class="text-2xl sm:text-3xl font-bold text-text-light mb-3 sm:mb-4 leading-tight">
				{heading500}
			</h1>
			<p class="text-base sm:text-lg text-text-muted leading-relaxed mb-8 max-w-prose mx-auto">
				{message500}
			</p>
			<a href="/" class="btn-primary inline-block px-8 py-3 font-semibold rounded-xl">{goHomeText}</a>
		</div>
	{:else if isNotFound && variant404 != null}
		<div
			class="error-404-full relative isolate flex w-full flex-1 flex-col rounded-none"
			data-404-variant={variant404}
		>
			<div
				class="error-404-art relative flex flex-none shrink-0 items-center justify-center px-3 pt-2 pb-1 min-[400px]:px-4 min-h-[22vh] min-[375px]:min-h-[26vh] sm:min-h-[32vh] md:min-h-[38vh] max-h-[34vh] min-[375px]:max-h-[38vh] sm:max-h-[42vh] md:max-h-[46vh]"
			>
				{#if variant404 === 'fr'}
					<span class="error-404-wave inline-flex max-h-full max-w-full items-center justify-center">
						<img
							src={img404}
							alt={alt404}
							width="200"
							height="200"
							class="error-404-illu drop-shadow-[0_10px_32px_rgba(0,0,0,0.4)]"
						/>
					</span>
				{:else}
					<img
						src={img404}
						alt={alt404}
						width="200"
						height="200"
						class="error-404-illu drop-shadow-[0_10px_32px_rgba(0,0,0,0.4)]"
					/>
				{/if}
			</div>

			<div
				class="error-404-copy relative mx-auto w-full max-w-xl shrink-0 px-4 text-center text-balance min-[400px]:px-5 sm:px-8"
			>
				<p
					class="text-text-light text-lg font-bold leading-snug tracking-tight min-[400px]:text-xl sm:text-2xl md:text-3xl"
				>
					{line1404}
				</p>
				{#if line2404}
					<p
						class="text-text-muted mt-2.5 text-base leading-relaxed min-[400px]:mt-3 sm:mt-3 sm:text-lg md:text-xl"
					>
						{line2404}
					</p>
				{/if}
			</div>

			<div
				class="error-404-cta relative mt-auto flex flex-col items-stretch justify-end px-4 pb-[max(1.25rem,env(safe-area-inset-bottom,0px))] pt-8 min-[400px]:px-5 sm:items-center sm:px-6 sm:pb-10 sm:pt-12 md:pt-16"
			>
				<a
					href="/"
					class="btn-primary w-full max-w-md self-center py-3.5 text-center text-sm font-semibold min-[400px]:text-base sm:w-auto sm:px-10 sm:py-3.5 sm:text-lg"
					>{button404}</a>
			</div>
		</div>
	{:else}
		<div class="text-center max-w-lg mx-auto">
			<h1 class="text-6xl font-bold text-text-light mb-4">{status}</h1>
			<p class="text-xl text-text-muted mb-8">{errorMessage}</p>
			<a href="/" class="btn-primary">{goHomeText}</a>
		</div>
	{/if}
</div>

<style>
	@keyframes gentle-wave {
		0%,
		100% {
			transform: rotate(-4deg) translateY(0);
		}
		50% {
			transform: rotate(4deg) translateY(-4px);
		}
	}

	.error-404-full {
		position: relative;
		box-sizing: border-box;
		width: 100%;
		max-width: 100%;
		min-height: calc(100vh - 4rem);
		min-height: calc(100dvh - 4rem);
		/* Breathing room below the app header on phones (~10–18vh, capped) */
		padding-top: max(env(safe-area-inset-top, 0px), clamp(1.25rem, 14vh, 6.5rem));
		padding-left: env(safe-area-inset-left, 0px);
		padding-right: env(safe-area-inset-right, 0px);
		overflow-x: hidden;
		overflow-y: auto;
		-webkit-overflow-scrolling: touch;
		background-image: linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
		background-color: var(--color-bg);
	}

	/* Full-bleed without 100vw on narrow viewports (avoids horizontal scroll with scrollbar) */
	@media (min-width: 641px) {
		.error-404-full {
			left: 50%;
			right: 50%;
			margin-left: -50vw;
			margin-right: -50vw;
			width: 100vw;
			max-width: 100vw;
			padding-top: max(env(safe-area-inset-top, 0px), clamp(1rem, 7vh, 4rem));
		}
	}

	.error-404-illu {
		width: auto;
		height: auto;
		max-width: min(16.5rem, 82vw);
		max-height: min(28vh, 220px);
		object-fit: contain;
	}

	@media (min-width: 400px) {
		.error-404-illu {
			max-width: min(17.5rem, 78vw);
			max-height: min(30vh, 240px);
		}
	}

	@media (min-width: 640px) {
		.error-404-illu {
			max-width: min(19rem, 72vw);
			max-height: min(36vh, 300px);
			transform: scale(1.06);
		}
	}

	@media (min-width: 768px) {
		.error-404-illu {
			max-width: min(20rem, 50vw);
			max-height: min(38vh, 360px);
			transform: scale(1.1);
		}
	}

	.error-404-wave {
		animation: gentle-wave 2.8s ease-in-out infinite;
	}

	@media (prefers-reduced-motion: reduce) {
		.error-404-wave {
			animation: none;
		}
	}

	/* --- Per-variant atmosphere (dark) --- */
	.error-404-full[data-404-variant='fr'] {
		background-image:
			radial-gradient(ellipse 100% 80% at 20% 15%, rgba(59, 130, 246, 0.16), transparent 58%),
			radial-gradient(ellipse 90% 60% at 85% 85%, rgba(239, 68, 68, 0.1), transparent 55%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}

	.error-404-full[data-404-variant='us'] {
		background-image:
			radial-gradient(ellipse 110% 85% at 12% 20%, rgba(239, 68, 68, 0.18), transparent 55%),
			radial-gradient(ellipse 100% 70% at 88% 78%, rgba(37, 99, 235, 0.14), transparent 52%),
			radial-gradient(ellipse 80% 50% at 50% 100%, rgba(248, 250, 252, 0.06), transparent 45%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}

	.error-404-full[data-404-variant='ja'] {
		background-image:
			radial-gradient(ellipse 95% 75% at 50% 10%, rgba(244, 114, 182, 0.14), transparent 50%),
			radial-gradient(ellipse 70% 55% at 80% 90%, rgba(251, 207, 232, 0.08), transparent 50%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}

	.error-404-full[data-404-variant='ca'] {
		background-image:
			radial-gradient(ellipse 100% 70% at 15% 30%, rgba(248, 113, 113, 0.12), transparent 55%),
			radial-gradient(ellipse 90% 65% at 85% 25%, rgba(56, 189, 248, 0.12), transparent 52%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}

	.error-404-full[data-404-variant='it'] {
		background-image:
			radial-gradient(ellipse 85% 70% at 18% 22%, rgba(34, 197, 94, 0.14), transparent 52%),
			radial-gradient(ellipse 75% 55% at 82% 80%, rgba(248, 113, 113, 0.1), transparent 50%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}

	.error-404-full[data-404-variant='gb'] {
		background-image:
			radial-gradient(ellipse 100% 75% at 25% 18%, rgba(59, 130, 246, 0.12), transparent 55%),
			radial-gradient(ellipse 80% 60% at 75% 85%, rgba(239, 68, 68, 0.09), transparent 48%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}

	.error-404-full[data-404-variant='de'] {
		background-image:
			radial-gradient(ellipse 110% 80% at 40% 25%, rgba(74, 222, 128, 0.1), transparent 55%),
			radial-gradient(ellipse 95% 70% at 70% 75%, rgba(100, 116, 139, 0.22), transparent 55%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}

	.error-404-full[data-404-variant='au'] {
		background-image:
			radial-gradient(ellipse 100% 75% at 20% 20%, rgba(245, 158, 11, 0.16), transparent 52%),
			radial-gradient(ellipse 90% 65% at 85% 70%, rgba(14, 165, 233, 0.12), transparent 50%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}

	.error-404-full[data-404-variant='ru'] {
		background-image:
			radial-gradient(ellipse 100% 85% at 50% 40%, rgba(148, 163, 184, 0.18), transparent 55%),
			radial-gradient(ellipse 60% 45% at 10% 90%, rgba(239, 68, 68, 0.06), transparent 45%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}

	.error-404-full[data-404-variant='es'] {
		background-image:
			radial-gradient(ellipse 105% 80% at 55% 15%, rgba(251, 146, 60, 0.16), transparent 52%),
			radial-gradient(ellipse 80% 55% at 15% 80%, rgba(234, 179, 8, 0.1), transparent 48%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}

	.error-404-full[data-404-variant='ch'] {
		background-image:
			radial-gradient(ellipse 90% 70% at 50% 20%, rgba(234, 179, 8, 0.14), transparent 50%),
			radial-gradient(ellipse 100% 75% at 50% 100%, rgba(100, 116, 139, 0.15), transparent 55%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}

	.error-404-full[data-404-variant='br'] {
		background-image:
			radial-gradient(ellipse 100% 75% at 30% 25%, rgba(34, 197, 94, 0.15), transparent 52%),
			radial-gradient(ellipse 85% 60% at 80% 70%, rgba(250, 204, 21, 0.14), transparent 48%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}

	.error-404-full[data-404-variant='ie'] {
		background-image:
			radial-gradient(ellipse 95% 70% at 22% 22%, rgba(22, 163, 74, 0.14), transparent 52%),
			radial-gradient(ellipse 85% 60% at 78% 75%, rgba(180, 83, 9, 0.12), transparent 48%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}

	.error-404-full[data-404-variant='nordic'] {
		background-image:
			radial-gradient(ellipse 110% 85% at 50% 30%, rgba(56, 189, 248, 0.12), transparent 55%),
			radial-gradient(ellipse 100% 70% at 20% 85%, rgba(100, 116, 139, 0.2), transparent 52%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}

	/* --- Light theme: softer washes --- */
	:global(:root.light) .error-404-full[data-404-variant='fr'] {
		background-image:
			radial-gradient(ellipse 100% 80% at 20% 15%, rgba(59, 130, 246, 0.1), transparent 58%),
			radial-gradient(ellipse 90% 60% at 85% 85%, rgba(239, 68, 68, 0.06), transparent 55%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}
	:global(:root.light) .error-404-full[data-404-variant='us'] {
		background-image:
			radial-gradient(ellipse 110% 85% at 12% 20%, rgba(239, 68, 68, 0.1), transparent 55%),
			radial-gradient(ellipse 100% 70% at 88% 78%, rgba(37, 99, 235, 0.08), transparent 52%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}
	:global(:root.light) .error-404-full[data-404-variant='ja'] {
		background-image:
			radial-gradient(ellipse 95% 75% at 50% 10%, rgba(244, 114, 182, 0.09), transparent 50%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}
	:global(:root.light) .error-404-full[data-404-variant='ca'] {
		background-image:
			radial-gradient(ellipse 100% 70% at 15% 30%, rgba(248, 113, 113, 0.07), transparent 55%),
			radial-gradient(ellipse 90% 65% at 85% 25%, rgba(56, 189, 248, 0.08), transparent 52%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}
	:global(:root.light) .error-404-full[data-404-variant='it'] {
		background-image:
			radial-gradient(ellipse 85% 70% at 18% 22%, rgba(34, 197, 94, 0.09), transparent 52%),
			radial-gradient(ellipse 75% 55% at 82% 80%, rgba(248, 113, 113, 0.06), transparent 50%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}
	:global(:root.light) .error-404-full[data-404-variant='gb'] {
		background-image:
			radial-gradient(ellipse 100% 75% at 25% 18%, rgba(59, 130, 246, 0.07), transparent 55%),
			radial-gradient(ellipse 80% 60% at 75% 85%, rgba(239, 68, 68, 0.05), transparent 48%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}
	:global(:root.light) .error-404-full[data-404-variant='de'] {
		background-image:
			radial-gradient(ellipse 110% 80% at 40% 25%, rgba(74, 222, 128, 0.07), transparent 55%),
			radial-gradient(ellipse 95% 70% at 70% 75%, rgba(100, 116, 139, 0.12), transparent 55%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}
	:global(:root.light) .error-404-full[data-404-variant='au'] {
		background-image:
			radial-gradient(ellipse 100% 75% at 20% 20%, rgba(245, 158, 11, 0.11), transparent 52%),
			radial-gradient(ellipse 90% 65% at 85% 70%, rgba(14, 165, 233, 0.07), transparent 50%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}
	:global(:root.light) .error-404-full[data-404-variant='ru'] {
		background-image:
			radial-gradient(ellipse 100% 85% at 50% 40%, rgba(148, 163, 184, 0.12), transparent 55%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}
	:global(:root.light) .error-404-full[data-404-variant='es'] {
		background-image:
			radial-gradient(ellipse 105% 80% at 55% 15%, rgba(251, 146, 60, 0.1), transparent 52%),
			radial-gradient(ellipse 80% 55% at 15% 80%, rgba(234, 179, 8, 0.07), transparent 48%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}
	:global(:root.light) .error-404-full[data-404-variant='ch'] {
		background-image:
			radial-gradient(ellipse 90% 70% at 50% 20%, rgba(234, 179, 8, 0.09), transparent 50%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}
	:global(:root.light) .error-404-full[data-404-variant='br'] {
		background-image:
			radial-gradient(ellipse 100% 75% at 30% 25%, rgba(34, 197, 94, 0.09), transparent 52%),
			radial-gradient(ellipse 85% 60% at 80% 70%, rgba(250, 204, 21, 0.09), transparent 48%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}
	:global(:root.light) .error-404-full[data-404-variant='ie'] {
		background-image:
			radial-gradient(ellipse 95% 70% at 22% 22%, rgba(22, 163, 74, 0.08), transparent 52%),
			radial-gradient(ellipse 85% 60% at 78% 75%, rgba(180, 83, 9, 0.07), transparent 48%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}
	:global(:root.light) .error-404-full[data-404-variant='nordic'] {
		background-image:
			radial-gradient(ellipse 110% 85% at 50% 30%, rgba(56, 189, 248, 0.08), transparent 55%),
			radial-gradient(ellipse 100% 70% at 20% 85%, rgba(100, 116, 139, 0.1), transparent 52%),
			linear-gradient(185deg, var(--color-bg) 0%, var(--color-bg-dark) 100%);
	}
</style>
