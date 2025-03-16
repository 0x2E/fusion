<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { getFavicon } from '$lib/api/favicon';
	import { logout } from '$lib/api/login';
	import type { Feed, Group } from '$lib/api/model';
	import {
		BookmarkCheck,
		CirclePlus,
		Inbox,
		List,
		LogOut,
		Settings,
		type Icon
	} from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { toggleShow } from './FeedActionAdd.svelte';
	import ThemeController from './ThemeController.svelte';

	interface Props {
		feeds: Feed[];
		groups: Group[];
	}

	let { feeds, groups }: Props = $props();

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
			await goto('/login');
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
			<li>
				<button
					onclick={() => {
						toggleShow();
					}}
					class="btn btn-sm bg-base-100 hover:bg-base-content/10"
				>
					<CirclePlus class="size-4" />
					<span>Add Feeds</span>
				</button>
			</li>
		</ul>

		<ul class="menu w-full font-medium">
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
							{#each feeds.filter((v) => v.group.id === group.id) as feed}
								{@const domain = new URL(feed.link).hostname}
								{@const textColor = feed.suspended
									? 'text-base-content/60'
									: feed.failure
										? 'text-error'
										: ''}
								<li>
									<a
										href="/feeds/{feed.id}"
										class={isHighlight('/feeds/' + feed.id) ? 'menu-active' : ''}
									>
										<div class="avatar">
											<div class="size-4 rounded-full">
												<img src={getFavicon(feed.link)} alt={feed.name} loading="lazy" />
											</div>
										</div>
										<span class={`line-clamp-1  ${textColor}`}>{feed.name}</span>
									</a>
								</li>
							{/each}
						</ul>
					</details>
				</li>
			{/each}
		</ul>
	</div>

	<div class="mt-8">
		<button onclick={handleLogout} class="btn btn-ghost btn-sm hover:text-error mt-auto w-full">
			<LogOut class="size-4" />
			Logout
		</button>
		<p class="text-base-content/60 text-center text-xs">
			<span>
				{version}.
			</span>
			<span>
				Logo by <a
					class="hover:underline"
					href="https://icons8.com/icon/FeQbTvGTsiN5/news"
					target="_blank"
				>
					Icons8
				</a>
			</span>
		</p>
	</div>
</div>
