<script lang="ts">
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';

	export let url: string;

	let copied = false;
	let timer: ReturnType<typeof setTimeout> | null = null;

	$: L = $locale;

	async function copy() {
		try {
			if (navigator.share && matchMedia('(pointer: coarse)').matches) {
				await navigator.share({ url });
				return;
			}
			await navigator.clipboard.writeText(url);
		} catch {
			return;
		}
		copied = true;
		if (timer) clearTimeout(timer);
		timer = setTimeout(() => (copied = false), 2000);
	}
</script>

<div class="flex flex-col sm:flex-row gap-2 items-stretch">
	<input
		type="text"
		readonly
		value={url}
		class="party-input flex-1 min-w-0 text-sm"
		aria-label={t('party.copy_link', undefined, L)}
		on:focus={(e) => e.currentTarget.select()}
	/>
	<button class="btn-secondary px-5 py-3 font-semibold whitespace-nowrap" on:click={copy}>
		{copied ? t('party.copied', undefined, L) : t('party.copy_link', undefined, L)}
	</button>
</div>
