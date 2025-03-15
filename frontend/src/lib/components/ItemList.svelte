<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { getFavicon } from '$lib/api/favicon';
	import { applyFilterToURL, parseURLtoFilter } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import ItemActionBookmark from './ItemActionBookmark.svelte';
	import ItemActionUnread from './ItemActionUnread.svelte';
	import ItemActionVisitLink from './ItemActionVisitLink.svelte';
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

	let filter = $derived(parseURLtoFilter(page.url.searchParams));
	async function handleChangePage(pageNumber: number) {
		filter.page = pageNumber;
		const url = page.url;
		applyFilterToURL(url, filter);
		await goto(url, { invalidateAll: true });
	}
</script>

<div>
	<ul data-sveltekit-preload-data="hover">
		{#each items as item}
			<li class="group rounded-md">
				<a
					href={'/items/' + item.id}
					class="hover:bg-base-300 relative flex w-full flex-col items-center justify-between space-y-1 space-x-2 rounded-md px-2 py-2 transition-colors md:flex-row"
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
							class="text-base-content/60 flex w-full justify-between gap-2 text-xs font-normal group-hover:hidden"
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
						class="invisible absolute right-1 w-fit justify-end gap-2 md:group-hover:visible md:group-hover:flex"
					>
						<ItemActionUnread data={item} />
						<ItemActionBookmark data={item} />
						<ItemActionVisitLink data={item} />
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
</div>
