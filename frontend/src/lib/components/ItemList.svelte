<script lang="ts">
	import { run } from 'svelte/legacy';

	import { Button } from './ui/button';
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
	import FeedsSelect from './FeedsSelect.svelte';
	import { Input } from './ui/input';
	import ItemActionVisitLink from './ItemActionVisitLink.svelte';
	import ItemActionBookmark from './ItemActionBookmark.svelte';
	import ItemActionUnread from './ItemActionUnread.svelte';

	interface Props {
		data: { feeds: Feed[]; items: { total: number; data: Item[] } };
	}

	let { data }: Props = $props();
	let filter = $state(parseURLtoFilter($page.url.searchParams));

	// NOTE: Svelte treats object as dirty, it may cause poorly reactive updates
	// when using it in two-way binding.
	// Therefore, we create an oldFilter as a control. Update url search params
	// only when the filter is NOT EQUAL to oldFilter.
	// TODO: this should be refactored after Svelte 5.0:
	// https://github.com/sveltejs/svelte/issues/4265#issuecomment-1812428837

	let oldFilter = Object.assign({}, filter);

	let selectedFeed = $state(filter?.feed_id);
	function updateSelectedFeed(id: number | undefined) {
		if (id === filter.feed_id) return;
		filter.feed_id = id !== -1 ? id : undefined;
		filter.page = 1;
		console.log(filter);
	}

	function setURLSearchParams(f: ListFilter) {
		console.log(
			`filter reactive updates:\nnew: ${JSON.stringify(f)}\nold: ${JSON.stringify(oldFilter)}`
		);

		let key: keyof ListFilter;
		let updated = false;
		for (key in f) {
			if (f[key] != oldFilter[key]) {
				updated = true;
				break;
			}
		}
		if (!updated) return;

		if (oldFilter.keyword !== filter.keyword) filter.page = 1;

		const p = new URLSearchParams($page.url.searchParams);
		for (key in f) {
			p.delete(key);
			if (f[key] !== undefined) {
				p.set(key, String(f[key]));
			}
		}

		oldFilter = Object.assign({}, filter);

		console.log(p.toString());
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
	function debounce(func: Function, wait: number): EventListener {
		let timeout: ReturnType<typeof setTimeout>;

		return function (this: HTMLElement, event: Event) {
			const context = this;

			const later = () => {
				func.apply(context, [event]);
			};

			clearTimeout(timeout);
			timeout = setTimeout(later, wait);
		};
	}

	const handleSearchInput = debounce(function (e: Event) {
		if (e.target instanceof HTMLInputElement) {
			filter.keyword = e.target.value;
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

	const actions: { icon: ComponentType<Icon>; tooltip: string; handler: () => void }[] = [
		{ icon: CheckCheckIcon, tooltip: 'Mark as Read', handler: handleMarkAllAsRead }
	];
	run(() => {
		updateSelectedFeed(selectedFeed);
	});
	run(() => {
		setURLSearchParams(filter);
	});
</script>

<div class="flex flex-col md:flex-row md:justify-between md:items-center w-full gap-2">
	<div class="flex flex-col md:flex-row gap-2">
		<FeedsSelect data={data.feeds} bind:selected={selectedFeed} className="w-full md:w-[200px]" />
		<Input
			type="search"
			placeholder="Search in title and content..."
			class="w-full md:w-[400px]"
			value={filter.keyword}
			on:input={handleSearchInput}
		/>
	</div>

	{#if data.items.data.length > 0}
		<div>
			{#each actions as action}
				<Tooltip.Root>
					<Tooltip.Trigger asChild >
						{#snippet children({ builder })}
												<Button
								builders={[builder]}
								on:click={action.handler}
								variant="outline"
								size="icon"
								class="w-full md:w-[40px]"
							>
								<action.icon size="20" />
								<span class="ml-1 md:hidden">{action.tooltip}</span>
							</Button>
																	{/snippet}
										</Tooltip.Trigger>
					<Tooltip.Content>
						<p>{action.tooltip}</p>
					</Tooltip.Content>
				</Tooltip.Root>
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
			bind:perPage={filter.page_size}
			bind:page={filter.page}
			
			
			class="w-auto mx-0"
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
								{/snippet}
				</Pagination.Root>

		<Select.Root
			items={[{ value: 10, label: '10' }]}
			onSelectedChange={(v) => {
				filter.page_size = v?.value || 10;
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
{/if}
