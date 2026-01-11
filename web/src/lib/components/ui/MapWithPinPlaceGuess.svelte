<script>
	import { onMount, createEventDispatcher } from 'svelte';
	import Button from './Button.svelte';
	
	let L = null;
	
	onMount(async () => {
		// Dynamically import Leaflet only on client side
		if (typeof window !== 'undefined') {
			L = (await import('leaflet')).default;
			initMap();
		}
	});
	
	export let lat = 0;
	export let lng = 0;
	export let zoom = 2;
	export let width = '100%';
	export let height = '400px';
	export let question = 'Click on the map to guess the location';
	export let confirmLabel = 'Confirm Guess';
	export let disabled = false;
	
	const dispatch = createEventDispatcher();
	
	let mapContainer;
	let map;
	let guessMarker = null;
	let guessLat = null;
	let guessLng = null;
	let hasGuess = false;
	
	// Custom pin icon for guess
	const createGuessPinIcon = (color = '#EF4444') => {
		if (!L) return null;
		return L.divIcon({
			className: 'custom-pin-icon',
			html: `
				<div style="position: relative;">
					<div style="width: 4px; height: 2px; background: rgba(0,0,0,0.2); border-radius: 50%; position: absolute; bottom: -2px; left: 50%; transform: translateX(-50%); filter: blur(2px);"></div>
					<svg width="32" height="40" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" style="filter: drop-shadow(0 2px 4px rgba(0,0,0,0.3));">
						<path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7z" fill="${color}" stroke="${color === '#EF4444' ? '#DC2626' : '#2563EB'}" stroke-width="1.5"/>
						<circle cx="12" cy="9" r="3" fill="white"/>
					</svg>
				</div>
			`,
			iconSize: [32, 40],
			iconAnchor: [16, 40]
		});
	};
	
	function initMap() {
		if (!mapContainer || !L) return;
		
		// Initialize map with dark theme
		map = L.map(mapContainer, {
			zoomControl: true,
			attributionControl: true
		});
		
		// Use dark tile layer
		L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
			attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
			subdomains: 'abcd',
			maxZoom: 19
		}).addTo(map);
		
		// Set initial view
		map.setView([lat || 20, lng || 0], zoom);
		
		// Handle map clicks
		map.on('click', (e) => {
			if (disabled) return;
			
			const { lat, lng } = e.latlng;
			guessLat = lat;
			guessLng = lng;
			hasGuess = true;
			
			// Remove previous marker
			if (guessMarker) {
				map.removeLayer(guessMarker);
			}
			
			// Add new marker
			const icon = createGuessPinIcon('#3B82F6');
			if (icon) {
				guessMarker = L.marker([lat, lng], { 
					icon,
					draggable: true
				}).addTo(map);
			} else {
				guessMarker = L.marker([lat, lng], { 
					draggable: true
				}).addTo(map);
			}
			
			// Allow dragging to adjust
			guessMarker.on('dragend', (e) => {
				const pos = e.target.getLatLng();
				guessLat = pos.lat;
				guessLng = pos.lng;
			});
		});
		
		return () => {
			if (map) {
				map.remove();
			}
		};
	}
	
	function handleConfirm() {
		if (!hasGuess || disabled) return;
		
		dispatch('guess', {
			lat: guessLat,
			lng: guessLng
		});
	}
	
	function handleClear() {
		if (guessMarker && map) {
			map.removeLayer(guessMarker);
			guessMarker = null;
		}
		hasGuess = false;
		guessLat = null;
		guessLng = null;
	}
</script>

<div class="map-guess-container">
	<div class="mb-3">
		<p class="text-text-light text-sm mb-2">{question}</p>
		{#if hasGuess}
			<div class="flex gap-2">
				<Button variant="primary" size="sm" on:click={handleConfirm} disabled={disabled}>
					{confirmLabel}
				</Button>
				<Button variant="secondary" size="sm" on:click={handleClear} disabled={disabled}>
					Clear
				</Button>
			</div>
		{/if}
	</div>
	<div class="map-container rounded-lg overflow-hidden border-2 border-white/20" style="width: {width}; height: {height};">
		<div bind:this={mapContainer} class="w-full h-full {disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-crosshair'}"></div>
	</div>
</div>

<style>
	:global(.map-container .leaflet-container) {
		background-color: var(--color-bg);
	}
	
	:global(.map-container .leaflet-control-zoom) {
		border: 1px solid rgba(255, 255, 255, 0.2);
		background-color: var(--color-surface);
		border-radius: 8px;
		overflow: hidden;
	}
	
	:global(:root.light .map-container .leaflet-control-zoom) {
		border-color: rgba(0, 0, 0, 0.2);
	}
	
	:global(.map-container .leaflet-control-zoom a) {
		background-color: var(--color-surface);
		color: var(--color-text);
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}
	
	:global(:root.light .map-container .leaflet-control-zoom a) {
		border-bottom-color: rgba(0, 0, 0, 0.1);
	}
	
	:global(.map-container .leaflet-control-zoom a:hover) {
		background-color: var(--color-surface-light);
		color: var(--color-text-light);
	}
</style>
