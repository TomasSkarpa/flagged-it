<script>
	// @ts-nocheck
	import { onMount, createEventDispatcher } from 'svelte';
	
	let L = null;
	
	onMount(async () => {
		// Dynamically import Leaflet only on client side
		if (typeof window !== 'undefined') {
			L = (await import('leaflet')).default;
			initMap();
		}
	});
	
	export let areaBoundary = null; // GeoJSON polygon or array of [lat, lng] points
	export let zoom = 2;
	export let width = '100%';
	export let height = '400px';
	export let showArea = true;
	export let disabled = false;
	
	const dispatch = createEventDispatcher();
	
	let mapContainer;
	let map;
	let areaLayer = null;
	let guessMarker = null;
	let hasGuess = false;
	let isWithinArea = false;
	
	// Check if point is within polygon (ray casting algorithm)
	function pointInPolygon(point, polygon) {
		const [x, y] = point;
		let inside = false;
		
		for (let i = 0, j = polygon.length - 1; i < polygon.length; j = i++) {
			const [xi, yi] = polygon[i];
			const [xj, yj] = polygon[j];
			
			const intersect = ((yi > y) !== (yj > y)) && (x < (xj - xi) * (y - yi) / (yj - yi) + xi);
			if (intersect) inside = !inside;
		}
		
		return inside;
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
		
		// Draw area boundary if provided
		if (areaBoundary && showArea) {
			let coordinates = [];
			
			// Handle GeoJSON format
			if (areaBoundary.type === 'Feature' || areaBoundary.type === 'FeatureCollection') {
				const features = areaBoundary.type === 'FeatureCollection' 
					? areaBoundary.features 
					: [areaBoundary];
				
				features.forEach(feature => {
					if (feature.geometry.type === 'Polygon') {
						coordinates = feature.geometry.coordinates[0].map(coord => [coord[1], coord[0]]); // GeoJSON is [lng, lat], Leaflet needs [lat, lng]
					}
				});
			} else if (Array.isArray(areaBoundary)) {
				// Handle array of [lat, lng] points
				coordinates = areaBoundary;
			}
			
			if (coordinates.length > 0) {
				areaLayer = L.polygon(coordinates, {
					color: '#3B82F6',
					fillColor: '#3B82F6',
					fillOpacity: 0.2,
					weight: 2,
					opacity: 0.7
				}).addTo(map);
				
				// Fit map to area
				const bounds = areaLayer.getBounds();
				if (bounds && bounds.isValid()) {
					map.fitBounds(bounds.pad(0.1));
				}
			}
		} else {
			// Default view
			map.setView([20, 0], zoom);
		}
		
		// Handle map clicks
		map.on('click', (e) => {
			if (disabled || hasGuess) return;
			
			const { lat, lng } = e.latlng;
			hasGuess = true;
			
			// Check if point is within area
			if (areaBoundary) {
				let polygon = [];
				
				// Extract polygon coordinates
				if (areaBoundary && areaBoundary.type === 'Feature' || areaBoundary && areaBoundary.type === 'FeatureCollection') {
					const features = areaBoundary.type === 'FeatureCollection' 
						? areaBoundary.features 
						: [areaBoundary];
					
					if (features[0] && features[0].geometry && features[0].geometry.type === 'Polygon') {
						polygon = features[0].geometry.coordinates[0].map((coord) => [coord[1], coord[0]]);
					}
				} else if (Array.isArray(areaBoundary)) {
					polygon = areaBoundary;
				}
				
				if (polygon.length > 0) {
					isWithinArea = pointInPolygon([lat, lng], polygon);
				}
			}
			
			// Add marker with color based on result
			const color = isWithinArea ? '#10B981' : '#EF4444';
			const strokeColor = isWithinArea ? '#059669' : '#DC2626';
			
			const icon = createPinIcon(color, strokeColor);
			if (icon) {
				guessMarker = L.marker([lat, lng], { icon }).addTo(map);
			} else {
				guessMarker = L.marker([lat, lng]).addTo(map);
			}
			
			// Dispatch event
			dispatch('guess', {
				lat,
				lng,
				withinArea: isWithinArea
			});
		});
		
		return () => {
			if (map) {
				map.remove();
			}
		};
	}
	
	// Update area when boundary changes
	$: if (map && areaBoundary && showArea && L) {
		if (areaLayer && map) {
			map.removeLayer(areaLayer);
		}
		
		let coordinates = [];
		if (areaBoundary && (areaBoundary.type === 'Feature' || areaBoundary.type === 'FeatureCollection')) {
			const features = areaBoundary.type === 'FeatureCollection' 
				? areaBoundary.features 
				: [areaBoundary];
			
			features.forEach(feature => {
				if (feature && feature.geometry && feature.geometry.type === 'Polygon') {
					coordinates = feature.geometry.coordinates[0].map(coord => [coord[1], coord[0]]);
				}
			});
		} else if (Array.isArray(areaBoundary)) {
			coordinates = areaBoundary;
		}
		
		if (coordinates.length > 0 && L && map) {
			areaLayer = L.polygon(coordinates, {
				color: '#3B82F6',
				fillColor: '#3B82F6',
				fillOpacity: 0.2,
				weight: 2,
				opacity: 0.7
			}).addTo(map);
			
			if (areaLayer) {
				map.fitBounds(areaLayer.getBounds().pad(0.1));
			}
		}
	}
</script>

<div class="map-area-container">
	{#if hasGuess}
		<div class="mb-3 p-3 rounded-lg border {isWithinArea ? 'bg-success/20 border-success' : 'bg-error/20 border-error'}">
			<p class="text-text-light text-sm font-semibold">
				{isWithinArea ? '✅ Within area!' : '❌ Outside area'}
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
