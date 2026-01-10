<script>
	import { onMount, createEventDispatcher } from 'svelte';
	
	let L = null;
	
	onMount(async () => {
		// Dynamically import Leaflet only on client side
		if (typeof window !== 'undefined') {
			L = (await import('leaflet')).default;
			initMap();
		}
	});
	
	export let correctLat = 0;
	export let correctLng = 0;
	export let zoom = 2;
	export let width = '100%';
	export let height = '400px';
	export let showCorrectAfterClick = true;
	export let disabled = false;
	
	const dispatch = createEventDispatcher();
	
	let mapContainer;
	let map;
	let correctMarker = null;
	let guessMarker = null;
	let distanceLine = null;
	let hasGuess = false;
	let distance = null;
	
	// Calculate distance between two points (Haversine formula)
	function calculateDistance(lat1, lng1, lat2, lng2) {
		const R = 6371; // Earth's radius in km
		const dLat = (lat2 - lat1) * Math.PI / 180;
		const dLng = (lng2 - lng1) * Math.PI / 180;
		const a = 
			Math.sin(dLat / 2) * Math.sin(dLat / 2) +
			Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) *
			Math.sin(dLng / 2) * Math.sin(dLng / 2);
		const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
		return R * c;
	}
	
	// Custom pin icons
	const createPinIcon = (color, strokeColor) => {
		if (!L) return null;
		return L.divIcon({
			className: 'custom-pin-icon',
			html: `
				<div style="position: relative;">
					<div style="width: 4px; height: 2px; background: rgba(0,0,0,0.2); border-radius: 50%; position: absolute; bottom: -2px; left: 50%; transform: translateX(-50%); filter: blur(2px);"></div>
					<svg width="32" height="40" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" style="filter: drop-shadow(0 2px 4px rgba(0,0,0,0.3));">
						<path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7z" fill="${color}" stroke="${strokeColor}" stroke-width="1.5"/>
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
		const centerLat = correctLat || 20;
		const centerLng = correctLng || 0;
		map.setView([centerLat, centerLng], zoom);
		
		// Handle map clicks
		map.on('click', (e) => {
			if (disabled || hasGuess) return;
			
			const { lat, lng } = e.latlng;
			hasGuess = true;
			
			// Add guess marker (blue)
			const guessIcon = createPinIcon('#3B82F6', '#2563EB');
			if (guessIcon) {
				guessMarker = L.marker([lat, lng], { icon: guessIcon }).addTo(map);
			} else {
				guessMarker = L.marker([lat, lng]).addTo(map);
			}
			
			// Show correct location if enabled
			if (showCorrectAfterClick && correctLat && correctLng) {
				const correctIcon = createPinIcon('#10B981', '#059669');
				if (correctIcon) {
					correctMarker = L.marker([correctLat, correctLng], { icon: correctIcon }).addTo(map);
				} else {
					correctMarker = L.marker([correctLat, correctLng]).addTo(map);
				}
				
				// Draw line between points
				distanceLine = L.polyline(
					[[lat, lng], [correctLat, correctLng]],
					{
						color: '#EF4444',
						weight: 3,
						opacity: 0.7,
						dashArray: '10, 10'
					}
				).addTo(map);
				
				// Calculate distance
				distance = calculateDistance(lat, lng, correctLat, correctLng);
				
				// Add distance label at midpoint
				const midLat = (lat + correctLat) / 2;
				const midLng = (lng + correctLng) / 2;
				const distKm = distance || 0;
				const distanceLabel = L.marker([midLat, midLng], {
					icon: L.divIcon({
						className: 'distance-label',
						html: `<div style="background: rgba(239, 68, 68, 0.9); color: white; padding: 4px 8px; border-radius: 4px; font-weight: bold; white-space: nowrap; border: 2px solid white;">${distKm.toFixed(0)} km</div>`,
						iconSize: [100, 30],
						iconAnchor: [50, 15]
					})
				}).addTo(map);
				
				// Fit bounds to show both markers
				if (guessMarker && correctMarker && distanceLine) {
					const group = new L.FeatureGroup([guessMarker, correctMarker, distanceLine, distanceLabel]);
					map.fitBounds(group.getBounds().pad(0.1));
				}
			}
			
			// Dispatch event with guess and distance
			dispatch('guess', {
				lat,
				lng,
				distance: distance || 0,
				correctLat,
				correctLng
			});
		});
		
		return () => {
			if (map) {
				map.remove();
			}
		};
	}
	
	// Update when correct location changes
	$: if (map && correctLat && correctLng && !hasGuess) {
		map.setView([correctLat, correctLng], zoom);
	}
</script>

<div class="map-compare-container">
	{#if hasGuess && distance !== null}
		<div class="mb-3 p-3 rounded-lg bg-surface border border-white/10">
			<p class="text-text-light text-sm">
				<span class="font-semibold">Distance:</span> {distance ? distance.toFixed(0) : '0'} km
			</p>
		</div>
	{/if}
	<div class="map-container rounded-lg overflow-hidden border-2 border-white/20" style="width: {width}; height: {height};">
		<div bind:this={mapContainer} class="w-full h-full {disabled || hasGuess ? 'opacity-50 cursor-not-allowed' : 'cursor-crosshair'}"></div>
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
	
	:global(.map-container .leaflet-control-zoom a) {
		background-color: var(--color-surface);
		color: var(--color-text);
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}
	
	:global(.map-container .leaflet-control-zoom a:hover) {
		background-color: var(--color-surface-light);
		color: var(--color-text-light);
	}
</style>
