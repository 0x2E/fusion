<script lang="ts">
	import moment from 'moment';
	import { Button } from './ui/button';
	import ItemAction from './ItemAction.svelte';
	import * as Select from '$lib/components/ui/select';
	import type { Item } from '$lib/api/model';

	export let data: Item[];
	$: allFeeds = getFeeds(data);
	let selectedFeed = 'all';
	$: filteredItems = filterFeed(data, selectedFeed);

	function getFeeds(allItems: Item[]) {
		const feeds = new Map<number, { id: number; name: string }>();
		allItems.map((v) => feeds.set(v.feed.id, v.feed));
		return Array.from(feeds.values());
	}

	function filterFeed(allItems: Item[], feedID: string) {
		if (feedID === 'all') return allItems;
		return allItems.filter((v) => v.feed.id === parseInt(feedID));
	}
</script>

<div>
	<Select.Root
		items={allFeeds.map((v) => {
			return { value: v.id.toString(), label: v.name };
		})}
		onSelectedChange={(v) => v && (selectedFeed = v.value)}
	>
		<Select.Trigger class="w-[180px]">
			<Select.Value placeholder="Filter by Feed" />
		</Select.Trigger>
		<Select.Content class="max-h-40 overflow-scroll">
			<Select.Item value="all">All Feeds</Select.Item>
			{#each allFeeds as feed}
				<Select.Item value={feed.id}>{feed.name}</Select.Item>
			{/each}
		</Select.Content>
	</Select.Root>
</div>
<ul class="mt-4">
	{#each filteredItems as item}
		<li class="group rounded-md">
			<Button
				href={'/items?id=' + item.id}
				class="flex justify-between items-center gap-8 py-6"
				variant="ghost"
			>
				<div class="w-3/4 truncate text-lg font-medium">
					{item.title}
				</div>
				<div class="flex justify-between items-center w-1/4">
					<div class="flex w-full justify-between text-sm text-muted-foreground group-hover:hidden">
						<div class="truncate">{item.feed.name}</div>
						<div class="truncate">
							{moment(item.pub_date).fromNow(true)}
						</div>
					</div>

					<div class="w-full hidden group-hover:inline-flex justify-end">
						<ItemAction data={{ id: item.id, link: item.link, unread: item.unread }} />
					</div>
				</div>
			</Button>
		</li>
	{:else}
		No data
	{/each}
</ul>
