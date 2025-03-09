<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { applyFilterToURL, parseURLtoFilter } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import ItemActionBookmark from './ItemActionBookmark.svelte';
	import ItemActionUnread from './ItemActionUnread.svelte';
	import Pagination from './Pagination.svelte';

	interface Props {
		total: number;
		items: Item[];
		highlightUnread?: boolean;
	}
	let { items, total, highlightUnread }: Props = $props();

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

	let filter = parseURLtoFilter(page.url.searchParams);
	async function handleChangePage(pageNumber: number) {
		filter.page = pageNumber;
		const url = page.url;
		applyFilterToURL(url, filter);
		goto(url, { invalidateAll: true });
	}
</script>

<ul data-sveltekit-preload-data="hover" class="px-4">
	{#each items as item}
		<li class="group rounded-md">
			<a href={'/items/' + item.id} class="btn btn-ghost flex items-center justify-between py-6">
				<div class="flex items-center gap-2">
					{#if highlightUnread}
						<div class="size-1.5">
							<div
								class={`bg-accent h-full w-full rounded-full ${item.unread ? '' : 'hidden'}`}
							></div>
						</div>
					{/if}
					<h2 class="truncate font-medium">
						{item.title}
					</h2>
				</div>
				<div class="flex w-1/3 flex-shrink-0 items-center justify-between md:w-1/4">
					<div
						class="text-base-content/60 flex w-full justify-end gap-2 text-xs font-normal group-hover:hidden"
					>
						<span class="w-full truncate">{item.feed.name}</span>
						<span class="w-[5ch] truncate">
							{timeDiff(item.pub_date)}
						</span>
					</div>

					<div class="hidden w-full justify-end gap-2 group-hover:inline-flex">
						<ItemActionUnread data={item} />
						<ItemActionBookmark data={item} />
					</div>
				</div>
			</a>
		</li>
	{:else}
		Nothing here.
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
