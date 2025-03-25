<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { listItems } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { defaultPageSize } from '$lib/consts';
	import { t } from '$lib/i18n';
	import { fullItemFilter } from '$lib/state.svelte';
	import { ChevronLeft, ChevronRight } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		itemID: number;
		index: number;
		action: 'next' | 'previous';
	}
	let { itemID, index: currentItemIndex, action }: Props = $props();

	async function getNextItem(): Promise<Item | null | undefined> {
		fullItemFilter.page = fullItemFilter.page ?? 1;
		fullItemFilter.page_size = fullItemFilter.page_size ?? defaultPageSize;

		let { total, items } = await listItems(fullItemFilter);
		if (total === 0) {
			return null;
		}

		let nextItemIndex = 0;

		if (action === 'previous') {
			nextItemIndex = currentItemIndex - 1;
		} else {
			if (itemID !== items[currentItemIndex].id) {
				// the current item has been filtered out,
				// and the item to its right has automatically filled the position.
				nextItemIndex = currentItemIndex;
			} else {
				nextItemIndex += currentItemIndex + 1;
			}
		}

		if (nextItemIndex >= 0 && nextItemIndex < defaultPageSize) {
			return items[nextItemIndex];
		}

		// turn the page
		if (nextItemIndex < 0) {
			fullItemFilter.page -= 1;
			if (fullItemFilter.page < 1) {
				fullItemFilter.page = 1;
				return null;
			}
			items = (await listItems(fullItemFilter)).items;
			return items.at(-1);
		}
		fullItemFilter.page += 1;
		items = (await listItems(fullItemFilter)).items;
		return items.at(0);
	}

	async function handleSwitchItem() {
		const filterBackup = Object.assign({}, fullItemFilter);
		try {
			const next = await getNextItem();
			if (!next) {
				toast.error(t('state.no_more_data'));
				Object.assign(fullItemFilter, filterBackup);
				return;
			}
			goto('/items/' + next.id, {
				invalidate: ['page:' + page.url.pathname]
			});
		} catch (e) {
			toast.error((e as Error).message);
			return;
		}
	}
</script>

<button
	onclick={handleSwitchItem}
	class={`btn lg:btn-ghost btn-circle lg:btn-xl fixed bottom-1 ${action === 'previous' ? 'left-1' : 'right-1'} lg:sticky lg:top-[50%]`}
>
	{#if action === 'previous'}
		<ChevronLeft />
	{:else}
		<ChevronRight />
	{/if}
</button>
