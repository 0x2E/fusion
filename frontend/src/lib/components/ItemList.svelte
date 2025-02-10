<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { page } from '$app/state';
	import { parseURLtoFilter, updateUnread } from '$lib/api/item';
	import type { Feed, Item } from '$lib/api/model';
	import * as Pagination from '$lib/components/ui/pagination';
	import * as Select from '$lib/components/ui/select';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { cn, debounce } from '$lib/utils';
	import { CheckCheck, type Icon as IconType } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import FeedsSelect from './FeedsSelect.svelte';
	import ItemActionBookmark from './ItemActionBookmark.svelte';
	import ItemActionUnread from './ItemActionUnread.svelte';
	import ItemActionVisitLink from './ItemActionVisitLink.svelte';
	import { Button, buttonVariants } from './ui/button';
	import { Input } from './ui/input';

	interface Props {
		data: { feeds: Feed[]; items: { total: number; data: Item[] } };
	}
	let { data }: Props = $props();

	let filter = $state(parseURLtoFilter(page.url.searchParams));
	function applyFilter() {
		console.log(`filter reactive updates:\nnew: ${JSON.stringify(filter)}`);

		const url = page.url;
		const p = url.searchParams;
		for (const [key, v] of Object.entries(filter)) {
			if (v !== undefined) {
				p.set(key, String(v));
			} else {
				p.delete(key);
			}
		}
		goto(url, { invalidateAll: true });
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

	const handleSearchInput = debounce(function (e: Event) {
		if (e.target instanceof HTMLInputElement) {
			filter.keyword = e.target.value;
			filter.page = 1;
			applyFilter();
		}
	}, 500);

	function fromNow(d: Date) {
		d = new Date(d);
		const now = new Date();
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

	const actions: { icon: typeof IconType; tooltip: string; handler: () => void }[] = [
		{ icon: CheckCheck, tooltip: 'Mark as Read', handler: handleMarkAllAsRead }
	];
</script>

<div class="flex flex-col md:flex-row md:justify-between md:items-center w-full gap-2">
	<div class="flex flex-col md:flex-row gap-2">
		<FeedsSelect
			data={data.feeds}
			selected={filter.feed_id}
			onSelectedChange={(id: number | undefined) => {
				filter.feed_id = id;
				filter.page = 1;
				applyFilter();
			}}
			className="w-full md:w-[200px]"
		/>
		<Input
			type="search"
			placeholder="Search in title and content..."
			value={filter.keyword}
			oninput={handleSearchInput}
			class="w-full md:w-[400px]"
		/>
	</div>

	{#if data.items.data.length > 0}
		<div>
			{#each actions as action}
				<Tooltip.Provider>
					<Tooltip.Root delayDuration={100}>
						<Tooltip.Trigger
							onclick={action.handler}
							class={cn(buttonVariants({ variant: 'outline', size: 'icon' }), 'w-full md:w-[40px]')}
						>
							<action.icon size="20" />
							<span class="ml-1 md:hidden">{action.tooltip}</span>
						</Tooltip.Trigger>
						<Tooltip.Content>
							{action.tooltip}
						</Tooltip.Content>
					</Tooltip.Root>
				</Tooltip.Provider>
			{/each}
		</div>
	{/if}
</div>

<ul data-sveltekit-preload-data="hover" class="mt-4">
	{#each data.items.data as item}
		<li class="group rounded-md">
			<Button
				href={'/items?id=' + item.id}
				class="flex justify-between items-center gap-2 py-6"
				variant="ghost"
			>
				<h2 class="truncate text-lg font-medium">
					{item.title}
				</h2>
				<div class="flex justify-between items-center flex-shrink-0 w-1/3 md:w-1/4">
					<div
						class="flex justify-end w-full gap-2 text-sm text-muted-foreground group-hover:hidden"
					>
						<span class="w-full truncate">{item.feed.name}</span>
						<span class="w-[5ch] truncate">
							{fromNow(item.pub_date)}
						</span>
					</div>

					<div class="w-full hidden group-hover:inline-flex justify-end">
						<ItemActionUnread data={item} buttonClass="hover:bg-gray-300 dark:hover:bg-gray-700" />
						<ItemActionBookmark
							data={item}
							buttonClass="hover:bg-gray-300 dark:hover:bg-gray-700"
						/>
						<ItemActionVisitLink
							data={item}
							buttonClass="hover:bg-gray-300 dark:hover:bg-gray-700"
						/>
					</div>
				</div>
			</Button>
		</li>
	{:else}
		No data
	{/each}
</ul>

{#if data.items.total > 1}
	<div class="flex flex-col sm:flex-row items-center justify-center mt-8 gap-2">
		<Pagination.Root
			count={data.items.total}
			perPage={filter.page_size}
			page={filter.page}
			onPageChange={(p) => {
				filter.page = p;
				applyFilter();
			}}
		>
			{#snippet children({ pages, currentPage })}
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
							<Pagination.Item isVisible={currentPage === page.value}>
								<Pagination.Link {page} isActive={currentPage === page.value}>
									{page.value}
								</Pagination.Link>
							</Pagination.Item>
						{/if}
					{/each}
					<Pagination.Item>
						<Pagination.NextButton />
					</Pagination.Item>
				</Pagination.Content>
			{/snippet}
		</Pagination.Root>

		<Select.Root
			type="single"
			value={String(filter.page_size)}
			onValueChange={(v) => {
				filter.page_size = parseInt(v) || 10;
				applyFilter();
			}}
		>
			<Select.Trigger class="w-[110px]">Page Size</Select.Trigger>
			<Select.Content>
				{#each [10, 25, 50, 100, 200, 500] as size}
					<Select.Item value={String(size)}>{size}</Select.Item>
				{/each}
			</Select.Content>
		</Select.Root>
	</div>
{/if}
