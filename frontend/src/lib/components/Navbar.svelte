<script lang="ts">
	import { page } from '$app/stores';
	import { Button } from '$lib/components/ui/button';
	import { MenuIcon, XIcon } from 'lucide-svelte';
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
	let showMenu = false;
</script>

<nav class="block w-full sm:mt-3 mb-6">
	<div class="flex flex-col items-center w-full sm:max-w-[500px] sm:mx-auto bg-background">
		<div
			class="flex justify-between sm:justify-around w-full px-2 sm:px-6 py-2 sm:py-4 sm:rounded-2xl shadow-md sm:border"
		>
			<img src="/icon-96.png" alt="icon" class="w-10" />
			<div class="hidden sm:block">
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
			<ThemeToggler className="hidden sm:flex" />
			<Button
				variant="ghost"
				size="icon"
				on:click={() => (showMenu = !showMenu)}
				class="flex sm:hidden"
			>
				{#if showMenu}
					<XIcon size="15" />
				{:else}
					<MenuIcon size="15" />
				{/if}
			</Button>
		</div>

		<div class="relative w-full">
			{#if showMenu}
				<div class="flex flex-col w-full absolute top-0 bg-background shadow-md">
					{#each links as l}
						<Button
							variant="ghost"
							href={l.url}
							on:click={() => (showMenu = false)}
							class="w-full {l.highlight ? 'bg-accent text-accent-foreground' : ''}"
						>
							{l.label}
						</Button>
					{/each}
					<ThemeToggler className="w-full" />
				</div>
			{/if}
		</div>
	</div>
</nav>
