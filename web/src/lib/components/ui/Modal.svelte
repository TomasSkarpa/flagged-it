<script>
	// @ts-nocheck
	export let open = false;
	export let title = '';
	export let onClose = null;
	
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
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm animate-fade-in"
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
		<div class="card-game max-w-md w-full animate-slide-up relative z-10" on:click|stopPropagation on:keydown|stopPropagation role="document">
			<div class="flex items-center justify-between mb-6">
				{#if title}
					<h2 id="modal-title" class="text-section text-sandy-light font-heading">{title}</h2>
				{/if}
				{#if onClose}
					<button
						on:click={handleClose}
						class="btn-icon p-2"
						aria-label="Close modal"
					>
						✕
					</button>
				{/if}
			</div>
			<div>
				<slot />
			</div>
		</div>
	</div>
{/if}

