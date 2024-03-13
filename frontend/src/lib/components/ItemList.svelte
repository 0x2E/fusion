<script lang="ts">
	import moment from 'moment';
	import { Button } from './ui/button';
	import ItemAction from './ItemAction.svelte';
	import * as Select from '$lib/components/ui/select';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import * as Pagination from '$lib/components/ui/pagination';
	import type { Feed, Item } from '$lib/api/model';
	import { type ListFilter, updateUnread, parseURLtoFilter } from '$lib/api/item';
	import { toast } from 'svelte-sonner';
	import { type ComponentType } from 'svelte';
	import { CheckCheckIcon, type Icon } from 'lucide-svelte';
	import { page } from '$app/stores';
	import { goto, invalidateAll } from '$app/navigation';

	export let data: { feeds: Feed[]; items: { total: number; data: Item[] } };
	data.items.data = data.items.data.sort(
		(a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
	);
	const filter = parseURLtoFilter($page.url.searchParams);

	type feedOption = { label: string; value: number };
	const defaultSelectedFeed: feedOption = { value: -1, label: 'All Feeds' };
	const allFeeds: feedOption[] = data.feeds
		.map((f) => {
			return { value: f.id, label: f.name };
		})
		.concat(defaultSelectedFeed)
		.sort((a, b) => a.value - b.value);
	let selectedFeed = allFeeds.find((v) => v.value === filter.feed_id) || defaultSelectedFeed;

	let currentPage = filter.page;
	let pageSize = filter.page_size;

	$: updateSelectedFeed(selectedFeed);
	function updateSelectedFeed(f: feedOption) {
		console.log(f);
		filter.feed_id = f.value !== -1 ? f.value : undefined;
		filter.page = 1;
		setURLSearchParams(filter);
	}

	$: updatePage(currentPage);
	function updatePage(p: number) {
		filter.page = p;
		setURLSearchParams(filter);
	}

	$: updatePageSize(pageSize);
	function updatePageSize(size: number) {
		if (size < 10 || size > 500) {
			toast.warning('Page size is unreasonable');
			return;
		}
		filter.page_size = size;
		filter.page = 1;
		setURLSearchParams(filter);
	}

	function setURLSearchParams(f: ListFilter) {
		const p = new URLSearchParams($page.url.searchParams);
		for (let key in f) {
			p.delete(key);
			if (f[key] !== undefined) {
				p.set(key, String(f[key]));
			}
		}
		goto('?' + p.toString());
	}

	async function handleMarkAllAsRead() {
		try {
			const ids = data.items.data.map((v) => v.id);
			await updateUnread(ids, false);
			toast.success('Update successfully');
			invalidateAll();
		} catch (e) {
			toast.error((e as Error).message);
		}
	}
	const actions: { icon: ComponentType<Icon>; tooltip: string; handler: () => void }[] = [
		{ icon: CheckCheckIcon, tooltip: 'Mark as Read', handler: handleMarkAllAsRead }
	];
</script>

<div class="flex justify-between items-center w-full">
	<Select.Root items={allFeeds} bind:selected={selectedFeed}>
		<Select.Trigger class="w-[180px]">
			<Select.Value placeholder="Filter by Feed" />
		</Select.Trigger>
		<Select.Content class="max-h-[200px] overflow-y-scroll">
			{#each allFeeds as feed}
				<Select.Item value={feed.value} class="truncate">{feed.label}</Select.Item>
			{/each}
		</Select.Content>
	</Select.Root>

	{#if data.items.data.length > 0}
		<div>
			{#each actions as action}
				<Tooltip.Root>
					<Tooltip.Trigger asChild let:builder>
						<Button builders={[builder]} variant="ghost" on:click={action.handler} size="icon">
							<svelte:component this={action.icon} size="20" />
						</Button>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<p>{action.tooltip}</p>
					</Tooltip.Content>
				</Tooltip.Root>
			{/each}
		</div>
	{/if}
</div>

<ul class="mt-4">
	{#each data.items.data as item}
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

<div class="flex flex-row sm:flex-row items-center justify-center mt-8 gap-2">
	<Pagination.Root
		count={data.items.total}
		perPage={filter.page_size}
		bind:page={currentPage}
		let:pages
		let:currentPage
		class="w-auto mx-0"
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

	<Select.Root
		items={[{ value: 10, label: '10' }]}
		onSelectedChange={(v) => {
			v && (pageSize = v.value);
		}}
	>
		<Select.Trigger class="w-[110px]">
			<Select.Value placeholder="Page Size" />
		</Select.Trigger>
		<Select.Content>
			{#each [10, 25, 50, 100, 200, 500] as size}
				<Select.Item value={size}>{size}</Select.Item>
			{/each}
		</Select.Content>
	</Select.Root>
</div>
