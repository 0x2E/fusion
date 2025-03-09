<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { logout } from '$lib/api/login';
	import type { Feed } from '$lib/api/model';
	import {
		BookmarkCheck,
		CircleAlert,
		Inbox,
		List,
		LogOut,
		Settings,
		type Icon
	} from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import ThemeController from './ThemeController.svelte';

	interface Props {
		feeds: Feed[];
	}

	let { feeds }: Props = $props();

	type GroupFeed = {
		id: number;
		name: string;
		feeds: Feed[];
	};

	let groups = $derived.by(() => {
		const map = new Map<number, GroupFeed>();
		feeds.forEach((f) => {
			let g = map.get(f.group.id);
			if (!g) {
				g = { id: f.group.id, name: f.group.name, feeds: [] };
				map.set(f.group.id, g);
			}
			g.feeds.push(f);
		});
		const groups: GroupFeed[] = [];
		for (const [_, g] of map) {
			// TODO sort feeds
			groups.push(g);
		}
		groups.sort((a, b) => a.id - b.id);
		return groups;
	});

	const version = import.meta.env.FUSION.version;

	type SystemNavLink = {
		label: string;
		url: string;
		icon: typeof Icon;
	};
	const systemLinks: SystemNavLink[] = [
		{ label: 'Unread', url: '/', icon: Inbox },
		{ label: 'Bookmark', url: '/bookmarks', icon: BookmarkCheck },
		{ label: 'All', url: '/all', icon: List },
		{ label: 'Settings', url: '/settings', icon: Settings }
	];

	function isHighlight(url: string): boolean {
		let chunks = page.url.pathname.split('/');
		while (chunks.length > 1) {
			if (chunks.join('/') === url) {
				return true;
			}
			chunks.pop();
		}
		return false;
	}

	async function handleLogout() {
		if (!confirm('Are you sure you want to log out?')) {
			return;
		}

		try {
			await logout();
			toast.success('Bye');
			goto('/login');
		} catch {
			toast.error('Failed to logout.');
		}
	}
</script>

<div class="flex h-full flex-col justify-between">
	<div>
		<div class="flex items-center justify-between gap-2">
			<a
				href="https://github.com/0x2E/fusion"
				target="_blank"
				class="btn btn-ghost hover:bg-base-content/10 flex items-center justify-start gap-2"
			>
				<img src="/icon-96.png" alt="icon" class="w-6" />
				<span class="text-lg font-bold">Fusion</span>
			</a>
			<ThemeController />
		</div>

		<ul class="menu mt-4 w-full font-medium">
			{#each systemLinks as v}
				<li>
					<a href={v.url} class={isHighlight(v.url) ? 'menu-active' : ''}>
						<v.icon class="size-4" /><span>{v.label}</span>
					</a>
				</li>
			{/each}
		</ul>

		<ul class="menu w-full">
			<li class="menu-title">Feeds</li>
			{#each groups as group, index}
				<li>
					<details open={index === 0}>
						<summary class="overflow-hidden">
							<span class="line-clamp-1">{group.name}</span>
						</summary>
						<ul>
							{#each group.feeds as feed}
								<li>
									<a
										href="/feeds/{feed.id}"
										class={isHighlight('/feeds/' + feed.id) ? 'menu-active' : ''}
									>
										{#if feed.failure}
											<CircleAlert class="text-error size-3" />
										{/if}
										<span class="line-clamp-1">{feed.name}</span>
									</a>
								</li>
							{/each}
						</ul>
					</details>
				</li>
			{/each}
		</ul>
	</div>

	<div>
		<button onclick={handleLogout} class="btn btn-soft btn-sm mt-auto w-full">
			<LogOut class="size-4" />
			Logout</button
		>
		<p class="text-base-content/60 mt-2 text-center text-xs">
			<span>
				{version}.
			</span>
			<span>
				Icon by <a href="https://icons8.com/icon/FeQbTvGTsiN5/news" target="_blank">Icons8</a>
			</span>
		</p>
	</div>
</div>
