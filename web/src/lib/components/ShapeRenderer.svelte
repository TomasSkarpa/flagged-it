<script lang="ts">
	import { onMount, afterUpdate } from 'svelte';
	import { geoPath, geoIdentity } from 'd3-geo';
	import type { GeoJSON } from '$lib/api/shapeGame';

	export let geoJson: GeoJSON;
	export let width = 400;
	export let height = 300;
	export let fillColor = 'currentColor';
	export let strokeColor = 'none';
	export let strokeWidth = 0;

	let pathData = '';

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

	function renderShape() {
		if (!geoJson || !geoJson.features || geoJson.features.length === 0) {
			pathData = '';
			return;
		}

		const originalFeature = geoJson.features[0];
		
		// Filter to only significant polygons
		const filteredGeometry = filterLargePolygons(originalFeature.geometry);
		const feature = {
			...originalFeature,
			geometry: filteredGeometry
		};
		
		// Add padding
		const padding = 20;
		
		// Use geoIdentity with fitExtent to fit the shape within the padded area
		// fitExtent takes a bounding box [[x0, y0], [x1, y1]] and fits the geometry to it
		const projection = geoIdentity()
			.reflectY(true) // Flip Y because SVG Y increases downward, but geo lat increases upward
			.fitExtent(
				[[padding, padding], [width - padding, height - padding]], 
				feature as any
			);

		// Create path generator
		const path = geoPath().projection(projection);
		
		// Generate path data
		pathData = path(feature as any) || '';
	}

	onMount(() => {
		renderShape();
	});

	afterUpdate(() => {
		renderShape();
	});
</script>

<svg 
	{width} 
	{height} 
	viewBox="0 0 {width} {height}"
	class="shape-renderer"
>
	{#if pathData}
		<path 
			d={pathData} 
			fill={fillColor}
			stroke={strokeColor}
			stroke-width={strokeWidth}
		/>
	{/if}
</svg>

<style>
	.shape-renderer {
		display: block;
	}
</style>
