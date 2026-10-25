import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		sveltekit({
			preprocess: vitePreprocess(),
			adapter: adapter({
				pages: 'dist',
				assets: 'dist',
				fallback: 'index.html'
			}),
			paths: {
				relative: true
			}
		})
	],
	server: {
		host: '127.0.0.1',
		port: 9245,
		strictPort: true
	},
	clearScreen: false
});
