import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';
import { viteSingleFile } from 'vite-plugin-singlefile';
import { defineConfig } from 'vite';
import path from 'node:path';

export default defineConfig({
	base: './',
	plugins: [svelte(), tailwindcss(), viteSingleFile()],
	resolve: {
		alias: {
			$lib: path.resolve('./src/lib')
		}
	}
});
