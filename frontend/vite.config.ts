import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';
import { viteSingleFile } from 'vite-plugin-singlefile';
import { defineConfig } from 'vite';
import { createRequire } from 'node:module';
import path from 'node:path';

const require = createRequire(import.meta.url);

export default defineConfig({
	base: './',
	plugins: [svelte(), tailwindcss(), viteSingleFile()],
	resolve: {
		alias: {
			$lib: path.resolve('./src/lib'),
			// use tldts' pre-minified ESM build; the default "module" entry pulls the
			// unminified suffix trie and roughly doubles the bundle size
			tldts: require.resolve('tldts/dist/index.esm.min.js')
		}
	}
});
