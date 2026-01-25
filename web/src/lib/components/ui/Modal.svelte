<script lang="ts">
	export let open = false;
	export let title = '';
	export let onClose: (() => void) | null = null;
	
	function handleBackdropClick(event) {
		if (onClose && typeof onClose === 'function') {
			onClose();
		}
	}
	
	function handleKeydown(event) {
		if (event.key === 'Escape' && onClose && typeof onClose === 'function') {
			onClose();
		}
	}
	
	function handleClose() {
		if (onClose && typeof onClose === 'function') {
			onClose();
		}
	}
</script>

{#if open}
	<div 
		class="modal-overlay fixed inset-0 z-50 flex items-start sm:items-center justify-center p-2 sm:p-4 bg-black/50 backdrop-blur-sm animate-fade-in overflow-y-auto"
		role="dialog"
		aria-modal="true"
		aria-labelledby="modal-title"
	>
		<!-- Backdrop button for accessibility -->
		<button
			class="absolute inset-0 w-full h-full cursor-default"
			on:click={handleBackdropClick}
			on:keydown={handleKeydown}
			aria-label="Close modal"
			tabindex="-1"
		></button>
		<!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
		<div class="modal-content card-game max-w-md w-full animate-slide-up relative z-10 my-auto sm:my-0" on:click|stopPropagation on:keydown|stopPropagation role="document">
			<div class="flex items-center justify-between mb-4 sm:mb-6 flex-shrink-0">
				{#if title}
					<h2 id="modal-title" class="text-xl sm:text-section text-sandy-light font-heading pr-2">{title}</h2>
				{/if}
				{#if onClose}
					<button
						on:click={handleClose}
						class="btn-icon p-2 flex-shrink-0"
						aria-label="Close modal"
					>
						✕
					</button>
				{/if}
			</div>
			<div class="modal-body overflow-y-auto max-h-[calc(100vh-8rem)] sm:max-h-[calc(100vh-12rem)]">
				<slot />
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-overlay {
		/* Ensure overlay is scrollable on mobile */
		-webkit-overflow-scrolling: touch;
	}

	.modal-content {
		/* Ensure modal doesn't exceed viewport */
		max-height: calc(100vh - 1rem);
		display: flex;
		flex-direction: column;
	}

	@media (min-width: 640px) {
		.modal-content {
			max-height: calc(100vh - 2rem);
		}
	}

	.modal-body {
		/* Make content scrollable if needed */
		-webkit-overflow-scrolling: touch;
		flex: 1;
		min-height: 0;
	}

	/* Ensure proper spacing on mobile */
	@media (max-width: 640px) {
		.modal-content :global(.space-y-6) {
			gap: 1rem;
		}

		.modal-content :global(.space-y-6 > *) {
			margin-bottom: 1rem;
		}

		.modal-content :global(.space-y-6 > *:last-child) {
			margin-bottom: 0;
		}
	}

	/* Make close button square (1:1 aspect ratio) */
	.modal-content :global(.btn-icon) {
		aspect-ratio: 1 / 1;
		border-radius: 0.5rem !important; /* rounded-lg instead of rounded-full */
		width: 2.5rem;
		height: 2.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
	}
</style>