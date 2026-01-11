import * as sitemap from 'super-sitemap';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ url }) => {
	return await sitemap.response({
		origin: `${url.protocol}//${url.host}`,
		excludeRoutePatterns: ['^/debug.*', '^/ui-kit.*'],
		defaultChangefreq: 'weekly',
		defaultPriority: 0.8,
		processPaths: (paths: sitemap.PathObj[]) => {
			return paths.map((path) => {
				// Homepage gets highest priority
				if (path.path === '/') {
					return {
						...path,
						priority: 1.0,
						changefreq: 'weekly'
					};
				}
				
				// Game pages get high priority
				if (
					path.path.includes('-game') ||
					path.path.includes('higher-lower') ||
					path.path.includes('capital-game') ||
					path.path.includes('flag-game')
				) {
					return {
						...path,
						priority: 0.9,
						changefreq: 'weekly'
					};
				}
				
				// Library and scoreboard pages
				if (path.path.includes('library') || path.path.includes('scoreboard')) {
					return {
						...path,
						priority: 0.7,
						changefreq: 'monthly'
					};
				}
				
				// Default values are already set via defaultChangefreq and defaultPriority
				return path;
			});
		}
	});
};
