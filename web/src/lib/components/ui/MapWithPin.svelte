<script>
	import { createEventDispatcher } from 'svelte';
	
	export let pinLat = 0;
	export let pinLng = 0;
	export let showPin = true;
	export let width = '100%';
	export let height = '400px';
	export let interactive = false;
	
	const dispatch = createEventDispatcher();
	
	// For demo purposes, using a placeholder map background
	// In production, this would use an actual map library (Leaflet, Mapbox, etc.)
	
	function handleMapClick(event) {
		if (!interactive) return;
		
		const rect = event.currentTarget.getBoundingClientRect();
		const x = event.clientX - rect.left;
		const y = event.clientY - rect.top;
		
		// Convert pixel coordinates to percentage
		const latPercent = (y / rect.height) * 100;
		const lngPercent = (x / rect.width) * 100;
		
		dispatch('pinClick', { lat: latPercent, lng: lngPercent });
	}
</script>

<div class="relative w-full rounded-lg overflow-hidden border-2 border-white/20 bg-surface" style="width: {width}; height: {height}; background: linear-gradient(135deg, var(--color-bg) 0%, var(--color-bg-light) 100%);">
	<!-- Placeholder map background - In production, replace with actual map -->
	<div class="absolute inset-0 opacity-30" style="background-image: url('data:image/svg+xml,%3Csvg width=\'100\' height=\'100\' xmlns=\'http://www.w3.org/2000/svg\'%3E%3Cpath d=\'M0 0h100v100H0z\' fill=\'%230A0E27\'/%3E%3Cpath d=\'M20 20h60v60H20z\' fill=\'%23111827\' stroke=\'%23F8FAFC\' stroke-width=\'0.5\'/%3E%3C/svg%3E'); background-size: 200px 200px;"></div>
	
	{#if showPin}
		<!-- Pin marker -->
		<div 
			class="absolute transform -translate-x-1/2 -translate-y-full cursor-pointer z-10"
			style="left: {pinLng}%; top: {pinLat}%;"
			role="button"
			tabindex="0"
			aria-label="Pin at coordinates"
		>
			<div class="flex flex-col items-center">
				<!-- Pin shadow -->
				<div class="w-4 h-2 bg-black/20 rounded-full blur-sm mb-[-8px]"></div>
				<!-- Pin icon -->
				<svg class="w-8 h-10 drop-shadow-lg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
					<path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7z" fill="var(--color-error)" stroke="var(--color-error-dark)" stroke-width="1.5"/>
					<circle cx="12" cy="9" r="3" fill="white"/>
				</svg>
			</div>
		</div>
	{/if}
	
	<!-- Clickable overlay for placing pins -->
	<div 
		class="absolute inset-0 {interactive ? 'cursor-crosshair' : ''}" 
		role="button" 
		tabindex="0" 
		aria-label="Click to place pin"
		on:click={handleMapClick}
		on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { handleMapClick(e); } }}
	></div>
</div>

