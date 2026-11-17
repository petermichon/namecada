import path from 'node:path';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';
import { viteSingleFile } from 'vite-plugin-singlefile';
import { defineConfig } from 'vite';

const api = { '/v1': 'http://127.0.0.1:8080' };

// Tauri's mobile CLI sets TAURI_DEV_HOST to the address the device will use
// (`tauri android dev --host …`), so the dev server must listen there.
const host = process.env.TAURI_DEV_HOST;

export default defineConfig({
	base: './',
	plugins: [svelte(), tailwindcss(), viteSingleFile()],
	resolve: {
		alias: {
			$lib: path.resolve('./src/lib')
		}
	},
	// The app talks to /v1/... on its own origin. In dev Vite proxies that to
	// `serve`; in production the same process serves the built app and the API.
	server: {
		host: host || false,
		port: 5221,
		strictPort: true,
		proxy: api
	},
	preview: { proxy: api }
});
