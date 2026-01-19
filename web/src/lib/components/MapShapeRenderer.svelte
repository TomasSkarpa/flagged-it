<script lang="ts">
	import { onMount, afterUpdate } from 'svelte';
	import { geoPath, geoMercator, geoBounds, geoDistance } from 'd3-geo';
	import { getWorldGeoJSON } from '$lib/api/debug';
	import type { GeoJSON } from '$lib/api/shapeGame';

	export let targetGeoJson: GeoJSON;
	export let mode: 'easy' | 'medium' = 'easy'; // 'easy' = world view, 'medium' = neighbors view
	export let width = 400;
	export let height = 300;
	export let targetFillColor = 'white';
	export let targetStrokeColor = 'none';
	export let targetStrokeWidth = 0;

	let worldGeoJson: GeoJSON | null = null;
	let worldPathData: string[] = [];
	let targetPathData = '';
	let isLoading = true;
	let error: string | null = null;

	// Calculate approximate area of a polygon ring using shoelace formula
	function calculateRingArea(ring: number[][]): number {
		let area = 0;
		for (let i = 0; i < ring.length - 1; i++) {
			area += ring[i][0] * ring[i + 1][1];
			area -= ring[i + 1][0] * ring[i][1];
		}
		return Math.abs(area) / 2;
	}

	// Filter MultiPolygon to only keep significant polygons
	function filterLargePolygons(geometry: any): any {
		if (geometry.type === 'Polygon') {
			return geometry;
		}
		
		if (geometry.type !== 'MultiPolygon') {
			return geometry;
		}

		const polygons = geometry.coordinates as number[][][][];
		
		// Calculate area for each polygon
		const polygonAreas = polygons.map((poly, idx) => ({
			index: idx,
			area: calculateRingArea(poly[0]), // Use outer ring
			coords: poly
		}));

		// Sort by area descending
		polygonAreas.sort((a, b) => b.area - a.area);

		// Get total area
		const totalArea = polygonAreas.reduce((sum, p) => sum + p.area, 0);
		
		// Keep polygons that together make up 95% of the area, or at least the top 10
		let cumulativeArea = 0;
		const significantPolygons: number[][][][] = [];
		
		for (const poly of polygonAreas) {
			// Skip polygons that cross the date line (have both very negative and very positive longitudes)
			const lons = poly.coords[0].map((c: number[]) => c[0]);
			const minLon = Math.min(...lons);
			const maxLon = Math.max(...lons);
			
			// Skip if polygon spans more than 180 degrees (date line issue)
			if (maxLon - minLon > 180) continue;
			
			// Skip very small polygons (less than 0.1% of total)
			if (poly.area < totalArea * 0.001 && significantPolygons.length >= 3) continue;
			
			significantPolygons.push(poly.coords);
			cumulativeArea += poly.area;
			
			// Stop if we have 95% of the area or 20 polygons
			if (cumulativeArea > totalArea * 0.95 || significantPolygons.length >= 20) break;
		}

		return {
			type: 'MultiPolygon',
			coordinates: significantPolygons
		};
	}

	// Get centroid of a feature for distance calculation
	function getFeatureCentroid(feature: any): [number, number] {
		const bounds = geoBounds(feature);
		return [
			(bounds[0][0] + bounds[1][0]) / 2,
			(bounds[0][1] + bounds[1][1]) / 2
		];
	}

	// Filter neighbors for medium mode
	function filterNeighbors(worldFeatures: any[], targetFeature: any): any[] {
		const targetCentroid = getFeatureCentroid(targetFeature);
		const targetBounds = geoBounds(targetFeature);
		
		// Expand bounds by a few degrees for neighbor detection
		const padding = 5; // degrees
		const expandedBounds = [
			[targetBounds[0][0] - padding, targetBounds[0][1] - padding],
			[targetBounds[1][0] + padding, targetBounds[1][1] + padding]
		];

		return worldFeatures.filter(f => {
			// Skip the target country itself
			if (f.id === targetFeature.id) return false;
			
			const featureBounds = geoBounds(f);
			const featureCentroid = getFeatureCentroid(f);
			
			// Check if feature overlaps with expanded bounds
			const overlaps = !(
				featureBounds[1][0] < expandedBounds[0][0] ||
				featureBounds[0][0] > expandedBounds[1][0] ||
				featureBounds[1][1] < expandedBounds[0][1] ||
				featureBounds[0][1] > expandedBounds[1][1]
			);
			
			if (!overlaps) return false;
			
			// Also check distance - keep features within reasonable distance
			const distance = geoDistance(targetCentroid, featureCentroid);
			return distance < 0.15; // radians, roughly 15 degrees
		});
	}

	function renderMap() {
		if (!targetGeoJson || !targetGeoJson.features || targetGeoJson.features.length === 0) {
			targetPathData = '';
			worldPathData = [];
			return;
		}

		if (mode === 'easy' && !worldGeoJson) {
			return; // Wait for world data to load
		}

		const targetFeature = targetGeoJson.features[0];
		const filteredTargetGeometry = filterLargePolygons(targetFeature.geometry);
		const filteredTargetFeature = {
			...targetFeature,
			geometry: filteredTargetGeometry
		};

		// Determine padding based on mode
		// More padding = more zoomed out = more context visible
		const padding = mode === 'easy' ? 80 : 60; // Medium mode (neighbors) needs more zoom-out too
		
		// Create projection and fit to target
		const projection = geoMercator()
			.fitExtent(
				[[padding, padding], [width - padding, height - padding]], 
				filteredTargetFeature as any
			);

		const path = geoPath().projection(projection);

		// Generate target path
		targetPathData = path(filteredTargetFeature as any) || '';

		// Generate world/neighbor paths
		worldPathData = [];
		
		if (mode === 'easy' && worldGeoJson) {
			// Easy mode: render all world features as grey dashed lines
			worldPathData = worldGeoJson.features
				.map(f => {
					const pathStr = path(f as any);
					return pathStr || '';
				})
				.filter(p => p !== '');
		} else if (mode === 'medium' && worldGeoJson) {
			// Medium mode: render only neighbors
			const neighbors = filterNeighbors(worldGeoJson.features, filteredTargetFeature);
			worldPathData = neighbors
				.map(f => {
					const pathStr = path(f as any);
					return pathStr || '';
				})
				.filter(p => p !== '');
		}
	}

	async function loadWorldGeo() {
		if (mode === 'easy' || mode === 'medium') {
			try {
				isLoading = true;
				error = null;
				worldGeoJson = await getWorldGeoJSON();
			} catch (err) {
				error = err instanceof Error ? err.message : 'Failed to load world GeoJSON';
				console.error('Failed to load world GeoJSON:', err);
			} finally {
				isLoading = false;
			}
		}
	}

	$: if (!worldGeoJson) {
		loadWorldGeo();
	}

	$: if (targetGeoJson && worldGeoJson) {
		renderMap();
	}

	onMount(() => {
		loadWorldGeo();
	});

	afterUpdate(() => {
		if (targetGeoJson && worldGeoJson) {
			renderMap();
		}
	});
</script>

<svg 
	{width} 
	{height} 
	viewBox="0 0 {width} {height}"
	class="map-shape-renderer"
>
	{#if isLoading}
		<text x={width / 2} y={height / 2} text-anchor="middle" class="loading-text">
			Loading...
		</text>
	{:else if error}
		<text x={width / 2} y={height / 2} text-anchor="middle" class="error-text">
			{error}
		</text>
	{:else}
		<!-- World/Neighbor features (grey dashed lines) -->
		{#each worldPathData as pathStr}
			<path 
				d={pathStr} 
				fill="none"
				stroke="#888"
				stroke-width="1"
				stroke-dasharray="4,2"
				opacity="0.7"
			/>
		{/each}
		
		<!-- Target country (white fill) -->
		{#if targetPathData}
			<path 
				d={targetPathData} 
				fill={targetFillColor}
				stroke={targetStrokeColor}
				stroke-width={targetStrokeWidth}
			/>
		{/if}
	{/if}
</svg>

<style>
	.map-shape-renderer {
		display: block;
		width: 100%;
		height: auto;
		max-width: 100%;
	}

	.loading-text,
	.error-text {
		fill: var(--color-text-muted, #888);
		font-size: 14px;
	}
</style>