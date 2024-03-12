<script lang="ts">
	import moment from 'moment';
	import { Button } from './ui/button';
	import ItemAction from './ItemAction.svelte';
	import * as Select from '$lib/components/ui/select';
	import * as Pagination from '$lib/components/ui/pagination';
	import type { Feed, Item } from '$lib/api/model';
	import { listItems, type ListFilter } from '$lib/api/item';
	import { toast } from 'svelte-sonner';
	import { allFeeds as fetchAllFeeds } from '$lib/api/feed';

	export let filter: ListFilter = { offset: 0, count: 10 };

	if (filter.offset === undefined) filter.offset = 0;
	if (filter.count === undefined) filter.count = 10;

	fetchAllFeeds()
		.then((v) => {
			allFeeds = v;
		})
		.catch((e) => {
			toast.error('Failed to fetch feeds data: ' + e);
		});

	let data: Item[] = [];
	let allFeeds: Feed[] = [];
	let currentPage = 1;
	let total = 0;

	$: filter.offset = (currentPage - 1) * (filter?.count || 10);
	$: fetchItems(filter);

	async function fetchItems(filter: ListFilter) {
		try {
			const resp = await listItems(filter);
			data = resp.items.sort(
				(a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
			);
			total = resp.total;
		} catch (e) {
			toast.error((e as Error).message);
		}
	}
</script>

<div>
	<Select.Root
		items={allFeeds.map((v) => {
			return { value: v.id.toString(), label: v.name };
		})}
		onSelectedChange={(v) => {
			if (!v) return;
			const feedID = parseInt(v.value);
			filter.feed_id = feedID > 0 ? feedID : undefined;
			filter.offset = 0;
		}}
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
	{#each data as item}
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
						<ItemAction bind:data={item} />
					</div>
				</div>
			</Button>
		</li>
	{:else}
		No data
	{/each}
</ul>

{#if total > (filter?.count || 10)}
	<Pagination.Root
		count={total}
		perPage={filter.count}
		bind:page={currentPage}
		let:pages
		let:currentPage
		class="mt-8"
	>
		<Pagination.Content class="flex-wrap">
			<Pagination.Item>
				<Pagination.PrevButton />
			</Pagination.Item>
			{#each pages as page (page.key)}
				{#if page.type === 'ellipsis'}
					<Pagination.Item>
						<Pagination.Ellipsis />
					</Pagination.Item>
				{:else}
					<Pagination.Item>
						<Pagination.Link {page} isActive={currentPage == page.value}>
							{page.value}
						</Pagination.Link>
					</Pagination.Item>
				{/if}
			{/each}
			<Pagination.Item>
				<Pagination.NextButton />
			</Pagination.Item>
		</Pagination.Content>
	</Pagination.Root>
{/if}
