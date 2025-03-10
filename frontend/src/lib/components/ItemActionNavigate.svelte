<script lang="ts">
	import { goto } from '$app/navigation';
	import { listItems } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { fullItemFilter } from '$lib/state.svelte';
	import { ChevronLeftIcon, ChevronRightIcon } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import ItemActionBase from './ItemActionBase.svelte';

	interface Props {
		data: Item;
		buttonClass?: string;
		iconSize?: number;
	}
	let { data, buttonClass = '', iconSize = 18 }: Props = $props();

	const itemFilter = Object.assign({}, fullItemFilter);
	let currentItemIndex = $state(0);
	let disabled = $state(false);

	onMount(async () => {
		const { items } = await listItems(itemFilter);
		currentItemIndex = items.findIndex((item) => item.id == data.id);
		disabled = currentItemIndex === -1;
	});

	async function getNextItem(action: 'next' | 'previous'): Promise<Item | null> {
		itemFilter.page = itemFilter.page ?? 1;
		itemFilter.page_size = itemFilter.page_size ?? 10;

		let { total, items } = await listItems(itemFilter);
		if (total === 0) {
			return null;
		}

		if (action === 'previous') {
			currentItemIndex -= 1;
		} else {
			let index = items.findIndex((v) => v.id === data.id);
			if (index === -1) {
				// the old item has been filtered out,
				// and the item to its right has automatically filled the position.
			} else {
				currentItemIndex += 1;
			}
		}

		if (currentItemIndex >= itemFilter.page_size) {
			itemFilter.page += 1;
			currentItemIndex = 0;
		} else if (currentItemIndex < 0) {
			itemFilter.page -= 1;
			if (itemFilter.page < 1) {
				toast.error('No more items');

				itemFilter.page = 1;
				currentItemIndex = 0;
				return null;
			}
			currentItemIndex = itemFilter.page_size - 1;
		}

		items = (await listItems(itemFilter)).items;
		if (items.length == 0) {
			return null;
		}
		return items[currentItemIndex];
	}

	async function handleSwitchItem(action: 'next' | 'previous') {
		const filterBackup = Object.assign({}, itemFilter);
		const indexBackup = currentItemIndex;
		const next = await getNextItem(action);
		if (!next) {
			toast.error('No more items');
			Object.assign(itemFilter, filterBackup);
			currentItemIndex = indexBackup;
			return;
		}
		goto('/items?id=' + next.id, { invalidateAll: true });
	}
</script>

<ItemActionBase
	fn={() => {
		handleSwitchItem('previous');
	}}
	tooltip="Previous item"
	{buttonClass}
	{iconSize}
	icon={ChevronLeftIcon}
	{disabled}
/>
<ItemActionBase
	fn={() => {
		handleSwitchItem('next');
	}}
	tooltip="Next item"
	{buttonClass}
	{iconSize}
	icon={ChevronRightIcon}
	{disabled}
/>
