<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { getFavicon } from '$lib/api/favicon';
	import { applyFilterToURL, parseURLtoFilter } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { t } from '$lib/i18n';
	import ItemActionBookmark from './ItemActionBookmark.svelte';
	import ItemActionUnread from './ItemActionUnread.svelte';
	import ItemActionVisitLink from './ItemActionVisitLink.svelte';
	import Pagination from './Pagination.svelte';
	import { hotkey, shortcuts } from './ShortcutHelpModal.svelte';

	interface Props {
		data: Promise<{
			total: number;
			items: Item[];
		}>;
		highlightUnread?: boolean;
	}
	let { data, highlightUnread }: Props = $props();

	let loading = $state(false);
	// make items reactive so we can display the updates without reloading the page
	let items = $state<Item[]>([]);
	let total = $state(0);
	$effect(() => {
		loading = true;
		data
			.then((v) => {
				items = v.items;
				total = v.total;
			})
			.finally(() => {
				loading = false;
			});
	});

	function timeDiff(d: Date) {
		d = new Date(d);
		const now = new Date();
		if (d.getTime() > now.getTime()) {
			return 'now';
		}
		const hours = Math.floor((now.getTime() - d.getTime()) / (1000 * 60 * 60));
		const days = Math.floor(hours / 24);
		const months = Math.floor(days / 30);
		const years = Math.floor(days / 365);
		if (years > 0) return years + 'y';
		if (months > 0) return months + 'm';
		if (days > 0) return days + 'd';
		if (hours > 0) return hours + 'h';
		return '?';
	}

	let filter = $derived(parseURLtoFilter(page.url.searchParams));
	async function handleChangePage(pageNumber: number) {
		filter.page = pageNumber;
		const url = page.url;
		applyFilterToURL(url, filter);
		await goto(url, { invalidate: ['page:' + page.url.pathname] });
	}

	let selectedItemIndex = $state(-1);
	$effect(() => {
		if (items) {
			selectedItemIndex = -1;
		}
	});
	function moveItem(direction: 'prev' | 'next') {
		if (items.length === 0) return;

		if (direction === 'prev') {
			selectedItemIndex -= 1;
			if (selectedItemIndex < 0) {
				selectedItemIndex = items.length - 1;
			}
		} else {
			selectedItemIndex += 1;
			selectedItemIndex %= items.length;
		}

		const el = document.getElementById(`item-${selectedItemIndex}`);
		if (el) {
			el.focus();
		}
	}
</script>

<div>
	{#if loading}
		<div class="flex flex-col gap-1">
			<div class="skeleton h-10 w-full rounded"></div>
			<div class="skeleton h-10 w-full rounded"></div>
			<div class="skeleton h-10 w-full rounded"></div>
			<div class="skeleton h-10 w-full rounded"></div>
			<div class="skeleton h-10 w-full rounded"></div>
		</div>
	{:else}
		<!-- shortcut -->
		<div class="hidden">
			<button onclick={() => moveItem('next')} use:hotkey={shortcuts.nextItem.keys}
				>{shortcuts.nextItem.desc}</button
			>
			<button onclick={() => moveItem('prev')} use:hotkey={shortcuts.prevItem.keys}
				>{shortcuts.prevItem.desc}</button
			>
		</div>

		<ul data-sveltekit-preload-data="hover">
			{#each items as item, i}
				<li class="rounded-md">
					<a
						id={'item-' + i}
						href={'/items/' + item.id}
						class="group hover:bg-base-200 relative flex w-full flex-col items-center justify-between space-y-1 space-x-2 rounded-md px-2 py-2 transition-colors focus:ring-2 md:flex-row"
					>
						<div class="flex w-full md:w-[80%] md:shrink-0">
							<h2
								class={`line-clamp-2 w-full truncate font-medium md:line-clamp-1 ${highlightUnread && !item.unread ? 'text-base-content/60' : ''}`}
							>
								{item.title || item.link}
							</h2>
						</div>
						<div class="flex w-full md:grow">
							<div
								class="text-base-content/60 flex w-full justify-between gap-2 text-xs font-normal group-hover:hidden group-focus:hidden"
							>
								<div class="flex grow items-center space-x-2 overflow-x-hidden">
									<div class="avatar">
										<div class="size-4 rounded-full">
											<img src={getFavicon(item.feed.link)} alt={item.feed.name} loading="lazy" />
										</div>
									</div>
									<span class="line-clamp-1">
										{item.feed.name}
									</span>
								</div>
								<span class="w-[4ch] shrink-0 truncate text-right">
									{timeDiff(item.pub_date)}
								</span>
							</div>
						</div>
						<div
							class="invisible absolute right-1 w-fit justify-end gap-2 md:group-hover:visible md:group-hover:flex md:group-focus:visible md:group-focus:flex"
						>
							<ItemActionUnread bind:item={items[i]} enableHotkey={i === selectedItemIndex} />
							<ItemActionBookmark bind:item={items[i]} enableHotkey={i === selectedItemIndex} />
							<ItemActionVisitLink {item} enableHotkey={i === selectedItemIndex} />
						</div>
					</a>
				</li>
			{:else}
				{t('state.no_data')}
			{/each}
		</ul>

		<div class="mt-6 flex w-full justify-center">
			<Pagination
				currentPage={filter.page}
				pageSize={filter.page_size}
				{total}
				onPageChange={handleChangePage}
			/>
		</div>
	{/if}
</div>
