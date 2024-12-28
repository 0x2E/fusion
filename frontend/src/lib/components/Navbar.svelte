<script lang="ts">
	import { run } from 'svelte/legacy';

	import { page } from '$app/stores';
	import { Button } from '$lib/components/ui/button';
	import { LogOutIcon, MenuIcon, XIcon } from 'lucide-svelte';
	import ThemeToggler from './ThemeToggler.svelte';
	import { logout } from '$lib/api/login';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';

	interface link {
		label: string;
		url: string;
		highlight?: boolean;
	}

	let links: link[] = $state([
		{ label: 'Unread', url: '/' },
		{ label: 'Bookmark', url: '/bookmarks' },
		{ label: 'All', url: '/all' },
		{ label: 'Feeds', url: '/feeds' }
	]);
	let showMenu = $state(false);

	let bodyOverflowDefault = document.body.style.overflow;
	function disableBodyScoll(showMenu: boolean) {
		document.body.style.overflow = showMenu ? 'hidden' : bodyOverflowDefault;
	}

	async function handleLogout() {
		try {
			await logout();
			toast.success('Bye');
			goto('/login');
		} catch {
			toast.error('Failed to logout.');
		}
	}
	run(() => {
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
	});
	run(() => {
		disableBodyScoll(showMenu);
	});
</script>

<nav class="block w-full sm:mt-3 mb-6">
	<div class="flex flex-col items-center w-full sm:max-w-[500px] sm:mx-auto bg-background">
		<div
			class="flex justify-between sm:justify-around w-full px-4 sm:px-6 py-2 sm:py-4 sm:rounded-2xl shadow sm:border"
		>
			<a href="/">
				<img src="/icon-96.png" alt="icon" class="w-10" />
			</a>
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
				on:click={() => {
					handleLogout();
					showMenu = false;
				}}
				class="hidden sm:flex"
			>
				<LogOutIcon class="h-[1rem] w-[1rem]" />
				<span class="sr-only">Logout</span>
			</Button>
			<Button
				variant="outline"
				size="icon"
				on:click={() => (showMenu = !showMenu)}
				class="flex sm:hidden"
			>
				{#if showMenu}
					<XIcon />
				{:else}
					<MenuIcon />
				{/if}
			</Button>
		</div>

		<div
			class={`${showMenu ? 'opacity-100 visible' : 'opacity-0 invisible'} sm:hidden w-full h-[calc(100dvh)] z-50 fixed top-0 pt-14 pointer-events-none transition-all duration-300 origin-center overflow-y-auto`}
		>
			<div class="flex flex-col w-full h-full bg-background pointer-events-auto pt-6 pb-10">
				{#each links as l}
					<Button
						variant="ghost"
						href={l.url}
						on:click={() => (showMenu = false)}
						class={`w-full text-lg h-14 ${l.highlight ? 'bg-accent text-accent-foreground' : ''}`}
					>
						{l.label}
					</Button>
				{/each}
				<ThemeToggler className="w-full h-14" />
				<Button
					variant="outline"
					on:click={() => {
						handleLogout();
						showMenu = false;
					}}
					class="w-[50%] text-lg h-14 mt-auto mx-auto text-destructive"
				>
					Logout
				</Button>
			</div>
		</div>
	</div>
</nav>
