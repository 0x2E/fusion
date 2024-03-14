<script lang="ts">
	import { page } from '$app/stores';
	import { Button } from '$lib/components/ui/button';
	import ThemeToggler from './ThemeToggler.svelte';
	interface link {
		label: string;
		url: string;
		highlight?: boolean;
	}
	let links: link[] = [
		{ label: 'Unread', url: '/' },
		{ label: 'Bookmark', url: '/bookmarks' },
		{ label: 'All', url: '/all' },
		{ label: 'Feeds', url: '/feeds' }
	];
	$: {
		let path = $page.url.pathname;
		for (const l of links) {
			l.highlight = false;
			let p = path.split('/');
			while (p.length > 1) {
				if (p.join('/') === l.url) {
					l.highlight = true;
					break;
				}
				p.pop();
			}
		}
		links = links;
	}
	// TODO: responsive navbar
</script>

<nav class="block w-full sm:mt-3 mb-6">
	<div
		class="flex justify-around items-center w-full sm:max-w-[500px] mx-auto px-6 py-4 sm:rounded-2xl shadow-md sm:border bg-background"
	>
		<div class="flex items-center">
			<img src="/favicon.png" alt="logo" class="w-10" />
			<!-- <h2 class="font-bold text-xl">Fusion</h2> -->
		</div>
		<div>
			{#each links as l}
				<Button
					variant="ghost"
					href={l.url}
					class={l.highlight ? 'bg-accent text-accent-foreground' : ''}
				>
					{l.label}
				</Button>
			{/each}
		</div>
		<ThemeToggler />
	</div>
</nav>
