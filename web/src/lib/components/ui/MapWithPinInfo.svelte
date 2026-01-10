<script>
	import { onMount } from 'svelte';
	
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
	export let title = '';
	
	let mapContainer;
	let map;
	let marker;
	
	// Custom pin icon
	const createPinIcon = () => {
		if (!L) return null;
		return L.divIcon({
			className: 'custom-pin-icon',
			html: `
				<div style="position: relative;">
					<div style="width: 4px; height: 2px; background: rgba(0,0,0,0.2); border-radius: 50%; position: absolute; bottom: -2px; left: 50%; transform: translateX(-50%); filter: blur(2px);"></div>
					<svg width="32" height="40" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" style="filter: drop-shadow(0 2px 4px rgba(0,0,0,0.3));">
						<path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7z" fill="#EF4444" stroke="#DC2626" stroke-width="1.5"/>
						<circle cx="12" cy="9" r="3" fill="white"/>
					</svg>
				</div>
			`,
			iconSize: [32, 40],
			iconAnchor: [16, 40],
			popupAnchor: [0, -40]
		});
	};
	
	function initMap() {
		if (!mapContainer || !L) return;
		
		// Initialize map with dark theme
		map = L.map(mapContainer, {
			zoomControl: true,
			attributionControl: true
		});
		
		// Use dark tile layer (CartoDB Dark Matter)
		L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
			attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
			subdomains: 'abcd',
			maxZoom: 19
		}).addTo(map);
		
		// Set view and add marker
		map.setView([lat, lng], zoom);
		
		const icon = createPinIcon();
		if (icon) {
			marker = L.marker([lat, lng], { icon }).addTo(map);
		} else {
			marker = L.marker([lat, lng]).addTo(map);
		}
		
		if (title) {
			marker.bindPopup(title).openPopup();
		}
		
		return () => {
			if (map) {
				map.remove();
			}
		};
	}
	
	// Update marker position when props change
	$: if (map && marker && lat && lng && L) {
		marker.setLatLng([lat, lng]);
		map.setView([lat, lng], zoom);
		if (title) {
			marker.bindPopup(title);
		}
	}
</script>

<div class="map-container rounded-lg overflow-hidden border-2 border-white/20" style="width: {width}; height: {height};">
	<div bind:this={mapContainer} class="w-full h-full"></div>
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
	
	:global(.map-container .leaflet-popup-content-wrapper) {
		background-color: var(--color-surface);
		color: var(--color-text);
		border-radius: 8px;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}
	
	:global(.map-container .leaflet-popup-tip) {
		background-color: var(--color-surface);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}
</style>
