<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { getFavicon } from '$lib/api/favicon';
	import { logout } from '$lib/api/login';
	import type { Feed, Group } from '$lib/api/model';
	import { t } from '$lib/i18n';
	import {
		BookmarkCheck,
		CircleEllipsis,
		CirclePlus,
		Command,
		Inbox,
		List,
		LogOut,
		Settings,
		type Icon
	} from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { toggleShow as toggleShowFeedImport } from './FeedActionImport.svelte';
	import {
		hotkey,
		shortcuts,
		toggleShow as toggleShowShortcutHelpModal
	} from './ShortcutHelpModal.svelte';
	import ThemeController from './ThemeController.svelte';

	interface Props {
		feeds: Promise<Feed[]>;
		groups: Promise<Group[]>;
	}

	let { feeds, groups }: Props = $props();

	let feedList = $derived.by(async () => {
		const [feedsData, groupsData] = await Promise.all([feeds, groups]);
		const groupFeeds: { id: number; name: string; feeds: (Feed & { indexInList: number })[] }[] =
			[];
		let curIndexInList = 0;
		groupsData.forEach((group) => {
			groupFeeds.push({
				id: group.id,
				name: group.name,
				feeds: feedsData
					.filter((feed) => feed.group.id === group.id)
					.sort((a, b) => a.name.localeCompare(b.name))
					.map((feed) => ({
						...feed,
						indexInList: curIndexInList++
					}))
			});
		});
		return groupFeeds;
	});
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

	let selectedFeedIndex = $state(-1);
	let selectedFeedGroupId = $state(-1);
	$effect(() => {
		feeds.then(() => {
			selectedFeedIndex = -1;
			selectedFeedGroupId = -1;
		});
	});
	async function moveFeed(direction: 'prev' | 'next') {
		const feedList = await feeds;

		if (feedList.length === 0) return;

		if (direction === 'prev') {
			selectedFeedIndex -= 1;
			if (selectedFeedIndex < 0) {
				selectedFeedIndex = feedList.length - 1;
			}
		} else {
			selectedFeedIndex += 1;
			selectedFeedIndex %= feedList.length;
		}

		const el = document.getElementById(`sidebar-feed-${selectedFeedIndex}`);
		if (el) {
			selectedFeedGroupId = parseInt(el.getAttribute('data-group-id') ?? '-1');
			el.focus();
			// focus twice because <details> element's opening delay blocks the focus when
			// we open a new group (<details>)
			setTimeout(() => {
				if (!el) return;
				el.focus();
			}, 30);
		}
	}
</script>

<div class="hidden">
	<button onclick={() => moveFeed('next')} use:hotkey={shortcuts.nextFeed.keys}
		>{shortcuts.nextFeed.desc}</button
	>
	<button onclick={() => moveFeed('prev')} use:hotkey={shortcuts.prevFeed.keys}
		>{shortcuts.prevFeed.desc}</button
	>
</div>

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
						toggleShowFeedImport();
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
			{#await feedList}
				<div class="skeleton bg-base-300 h-10"></div>
			{:then groupData}
				{#each groupData as group, groupIndex}
					<li>
						<details open={groupIndex === 0 || selectedFeedGroupId === group.id}>
							<summary class="overflow-hidden">
								<span class="line-clamp-1">{group.name}</span>
							</summary>
							<ul>
								{#each group.feeds as feed}
									{@const textColor = feed.suspended
										? 'text-neutral-content/60'
										: feed.failure
											? 'text-error'
											: ''}
									<li>
										<a
											id="sidebar-feed-{feed.indexInList}"
											data-group-id={group.id}
											href="/feeds/{feed.id}"
											class={`${isHighlight('/feeds/' + feed.id) ? 'menu-active' : ''} focus:ring-2`}
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
			{/await}
		</ul>
	</div>

	<div class="mt-8">
		<div class="dropdown dropdown-top dropdown-center w-full">
			<div tabindex="0" role="button" class="btn btn-sm w-full">
				<CircleEllipsis class="size-4" />
				More
			</div>
			<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
			<div tabindex="0" class="dropdown-content bg-base-100 rounded-box z-1 w-52 p-2 shadow-sm">
				<ul class="menu w-full p-0">
					<li>
						<button
							onclick={() => toggleShowShortcutHelpModal()}
							use:hotkey={shortcuts.showHelp.keys}
						>
							<Command class="size-4" />
							Keyboard shortcuts
						</button>
					</li>
					<li>
						<button onclick={handleLogout} class="hover:text-error w-full">
							<LogOut class="size-4" />
							{t('common.logout')}
						</button>
					</li>
				</ul>
				<div class="bg-base-200 mt-2 p-2">
					<p class="text-base-content/60 text-xs">
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
		</div>
	</div>
</div>
