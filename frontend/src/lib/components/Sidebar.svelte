<script lang="ts">
	import { page } from '$app/state';
	import type { Feed } from '$lib/api/model';
	import { BookmarkCheck, Inbox, List, Settings, type Icon } from 'lucide-svelte';

	interface Props {
		feeds: Feed[];
	}

	let { feeds }: Props = $props();

	type GroupFeed = {
		id: number;
		name: string;
		feeds: {
			id: number;
			name: string;
		}[];
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
</script>

<a
	href="https://github.com/0x2E/fusion"
	target="_blank"
	class="btn hover:bg-base-content/10 flex items-center gap-2 justify-start"
>
	<img src="/icon-96.png" alt="icon" class="w-6" />
	<span class="text-lg font-bold">Fusion</span>
	<span class="text-xs text-base-content/60 font-light">
		{version}
	</span>
</a>

<ul class="menu w-full mt-4">
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
	<li><a>All</a></li>
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
								<span class="line-clamp-1">{feed.name}</span>
							</a>
						</li>
					{/each}
				</ul>
			</details>
		</li>
	{/each}
</ul>
