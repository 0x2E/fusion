<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { getFavicon } from '$lib/api/favicon';
	import { logout } from '$lib/api/login';
	import type { Feed, Group } from '$lib/api/model';
	import { t } from '$lib/i18n';
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
	import { toggleShow } from './FeedActionImport.svelte';
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
		{ label: t('common.unread'), url: '/', icon: Inbox },
		{ label: t('common.bookmark'), url: '/bookmarks', icon: BookmarkCheck },
		{ label: t('common.all'), url: '/all', icon: List },
		{ label: t('common.settings'), url: '/settings', icon: Settings }
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
		if (!confirm(t('auth.logout.confirm'))) {
			return;
		}

		try {
			await logout();
			await goto('/login');
		} catch {
			toast.error(t('auth.logout.failed_message'));
		}
	}
</script>

<div class="flex h-full flex-col justify-between">
	<div>
		<div class="flex items-center justify-between gap-2">
			<a
				href="https://github.com/0x2E/fusion"
				target="_blank"
				class="btn btn-ghost flex items-center justify-start gap-2"
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
					class="btn btn-sm btn-ghost bg-base-100"
				>
					<CirclePlus class="size-4" />
					<span>{t('feed.import.title')}</span>
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
			<li class="menu-title">{t('common.feeds')}</li>
			{#each groups as group, index}
				<li>
					<details open={index === 0}>
						<summary class="overflow-hidden">
							<span class="line-clamp-1">{group.name}</span>
						</summary>
						<ul>
							{#each feeds
								.filter((v) => v.group.id === group.id)
								.sort((a, b) => a.name.localeCompare(b.name)) as feed}
								{@const textColor = feed.suspended
									? 'text-neutral-content/60'
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
			{t('common.logout')}
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
