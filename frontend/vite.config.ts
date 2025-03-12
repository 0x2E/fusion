import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import * as process from 'process';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	define: {
		'import.meta.env.FUSION': JSON.stringify({
			version: process.env.VITE_FUSION_VERSION || 'unknown-version'
		})
	},
	server: {
		proxy: {
			'/api': {
				target: 'http://localhost:8080',
				changeOrigin: true
			}
		}
	}
});
