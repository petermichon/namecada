import { defineConfig } from 'vite';

// The page itself is plain HTML and CSS: Vite only serves it with hot reload in
// development and hashes the stylesheet into dist/ for the nginx image. The
// port is fixed so the dev proxy can route a hostname to it (devhost:
// `http://site.namecada.localhost:7357`).
export default defineConfig({
	server: { port: 5222, strictPort: true },
	preview: { port: 5222, strictPort: true }
});
