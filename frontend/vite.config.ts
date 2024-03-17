import { sveltekit } from '@sveltejs/kit/vite';
import { execSync } from 'child_process';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	define: {
		'import.meta.env.FUSION': JSON.stringify({
			version:
				execSync('git describe --tags --abbrev=0').toString().trimEnd() ||
				execSync('git rev-parse --short HEAD').toString().trimEnd()
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
