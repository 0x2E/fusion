<script lang="ts">
	import { goto } from '$app/navigation';
	import type { Item } from '$lib/api/model';
	import type { ListFilter } from '$lib/api/item';
	import { listItems } from '$lib/api/item';
	import { ArrowUpIcon, ArrowLeftIcon, ArrowRightIcon } from 'lucide-svelte';
	import ItemActionBase from './ItemActionBase.svelte';
	import ItemActionBookmark from './ItemActionBookmark.svelte';
	import ItemActionUnread from './ItemActionUnread.svelte';
	import ItemActionVisitLink from './ItemActionVisitLink.svelte';
	import { Separator } from './ui/separator';

	interface Props {
		data: Item;
		fixed?: boolean;
	}

	let { data, fixed = true }: Props = $props();
	let filter: ListFilter|undefined = sessionStorage.getItem("filter");
	if (filter) {
		filter = JSON.parse(filter);
	}

	function handleScrollTop(e: Event) {
		e.preventDefault();
		document.body.scrollIntoView({ behavior: 'smooth' });
	}

	async function _findItem(next=true): Item {
		let items = await listItems(filter);
		const currentIndex = items.items.findIndex(item => item.id == data.id);
		let modifier;
		if (next) {
			// If out of items
			if (currentIndex >= items.total) {
				console.error("Deal with this error better");
				return data;
			}
			// If switch to next page
			if (currentIndex >= items.items.length - 1) {
				filter.page = filter.page+1;
				items = await listItems(filter);
				return items.items[0];
			}
			modifier = 1;
		} else {
			if (currentIndex <= 0) {
				if (filter.page == 1) {
					console.error("Deal with this error better too");
					return data;
				}
				filter.page = filter.page-1;
				items = await listItems(filter)
				return items.items[items.items.length - 1];
			}
			modifier = -1;
		}
		const newItem = items.items[currentIndex + modifier];
		return newItem;
	}

	async function anotherItem(next=true) {
		await goto("/items?id=" + (await _findItem(next)).id)
	}

	async function previousItem(e: Event) {
		anotherItem(false);
	}

	async function nextItem(e: Event) {
		anotherItem(true);
	}
</script>

<div class="{fixed ? 'fixed' : ''} bottom-2 left-0 right-0">
	<div
		class="flex flex-row justify-center items-center gap-2 rounded-full border shadow w-fit mx-auto bg-background px-6 py-2"
	>
		<ItemActionUnread {data} />
		<Separator orientation="vertical" class="h-5" />
		<ItemActionBookmark {data} />
		<Separator orientation="vertical" class="h-5" />
		<ItemActionVisitLink {data} />
		<Separator orientation="vertical" class="h-5" />
		<ItemActionBase fn={handleScrollTop} tooltip="Back to Top" icon={ArrowUpIcon} />
		{#if filter}
		<Separator orientation="vertical" class="h-5" />
		<ItemActionBase fn={previousItem} tooltip="Previous item" icon={ArrowLeftIcon} />
		<ItemActionBase fn={nextItem} tooltip="Next item" icon={ArrowRightIcon} />
		{/if}
	</div>
</div>
