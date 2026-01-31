import type { PageLoad } from './$types';

// Define metadata for each debug route
// Add new routes here with their title, description, and icon
// Routes without metadata will use auto-generated defaults
const routeMetadata: Record<string, { title: string; description: string; icon: string }> = {
	flags: {
		title: 'All Flags',
		description: 'Browse all country flags with navigation',
		icon: '🚩'
	},
	shapes: {
		title: 'All Shapes',
		description: 'Browse all country silhouettes',
		icon: '🗺️'
	},
	'country-names': {
		title: 'Country Names',
		description: 'View all country names in different languages',
		icon: '🌍'
	},
	'background-test': {
		title: 'Background Test',
		description: 'Test background styles',
		icon: '🎨'
	},
	multiplayer: {
		title: 'Multiplayer',
		description: 'Test multiplayer room creation and joining',
		icon: '🎮'
	}
};

// Helper function to format route name (e.g., "country-names" -> "Country Names")
function formatRouteName(name: string): string {
	return name
		.split('-')
		.map(word => word.charAt(0).toUpperCase() + word.slice(1))
		.join(' ');
}

// Discover all routes dynamically using Vite's glob
// This automatically finds all +page.svelte files in subdirectories
const debugRoutes = import.meta.glob('./*/+page.svelte', { eager: false });

export const load: PageLoad = async () => {
	// Extract route names from the glob pattern
	const routes = Object.keys(debugRoutes)
		.map(path => {
			// Extract route name from path like "./flags/+page.svelte"
			const match = path.match(/\.\/([^/]+)\/\+page\.svelte/);
			return match ? match[1] : null;
		})
		.filter((name): name is string => name !== null)
		.map(name => {
			// Use metadata if available, otherwise generate defaults
			const metadata = routeMetadata[name];
			return {
				path: name,
				route: `/debug/${name}`,
				title: metadata?.title || formatRouteName(name),
				description: metadata?.description || `Debug page for ${formatRouteName(name)}`,
				icon: metadata?.icon || '🔧'
			};
		})
		.sort((a, b) => a.title.localeCompare(b.title));

	return {
		routes
	};
};
