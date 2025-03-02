import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import * as process from 'process';

export default defineConfig({
	plugins: [sveltekit()],
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
