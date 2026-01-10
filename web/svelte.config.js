import adapterAuto from '@sveltejs/adapter-auto';
import adapterStatic from '@sveltejs/adapter-static';
import adapterVercel from '@sveltejs/adapter-vercel';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

const isDesktopBuild = process.env.BUILD_TARGET === 'desktop';
const isVercel = process.env.VERCEL === '1' || process.env.VERCEL_ENV;

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://kit.svelte.dev/docs/integrations
	// for more information about preprocessors
	preprocess: vitePreprocess(),

	kit: {
		// Use static adapter for Wails desktop builds, Vercel adapter for Vercel deployments, auto for others
		adapter: isDesktopBuild 
			? adapterStatic({
				pages: 'build',
				assets: 'build',
				fallback: 'index.html',
				precompress: false,
				strict: true
			})
			: isVercel
				? adapterVercel({
					runtime: 'nodejs24.x'
				})
				: adapterAuto()
	}
};

export default config;

