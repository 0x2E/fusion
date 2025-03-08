<script lang="ts">
	import type { Item } from '$lib/api/model';
	import ItemActionBookmark from './ItemActionBookmark.svelte';
	import ItemActionUnread from './ItemActionUnread.svelte';
	import ItemActionVisitLink from './ItemActionVisitLink.svelte';

	interface Props {
		items: Item[];
		highlightUnread?: boolean;
	}
	let { items, highlightUnread }: Props = $props();

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
</script>

<ul data-sveltekit-preload-data="hover" class="px-4">
	{#each items as item}
		<li class="group rounded-md">
			<a href={'/items/' + item.id} class="btn btn-ghost flex justify-between items-center py-6">
				<div class="flex items-center gap-2">
					{#if highlightUnread}
						<div class={`size-2 ${item.unread ? '' : 'hidden'}`}>
							<div class="bg-accent rounded-full w-full h-full"></div>
						</div>
					{/if}
					<h2 class="truncate font-medium">
						{item.title}
					</h2>
				</div>
				<div class="flex justify-between items-center flex-shrink-0 w-1/3 md:w-1/4">
					<div
						class="flex justify-end w-full gap-2 text-xs text-base-content/60 group-hover:hidden font-normal"
					>
						<span class="w-full truncate">{item.feed.name}</span>
						<span class="w-[5ch] truncate">
							{timeDiff(item.pub_date)}
						</span>
					</div>

					<div class="w-full hidden group-hover:inline-flex justify-end gap-2">
						<ItemActionUnread data={item} />
						<ItemActionBookmark data={item} />
						<ItemActionVisitLink data={item} />
					</div>
				</div>
			</a>
		</li>
	{:else}
		No data
	{/each}
</ul>
