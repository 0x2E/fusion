<script lang="ts">
	import { goto } from '$app/navigation';
	import { listItems } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { fullItemFilter } from '$lib/state.svelte';
	import { ArrowUpIcon, ChevronLeftIcon, ChevronRightIcon } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
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

	let nextItem: Item | null = $state(null);
	let prevItem: Item | null = $state(null);

	$effect(() => {
		console.log(data.id);
		updatePrevNextItem();
	});

	function handleScrollTop(e: Event) {
		e.preventDefault();
		document.body.scrollIntoView({ behavior: 'smooth' });
	}

	async function updatePrevNextItem() {
		const { items } = await listItems(fullItemFilter);
		if (items.length == 0) {
			return;
		}

		const currentIndex = items.findIndex((item) => item.id == data.id);
		prevItem = items[currentIndex - 1] || null;
		nextItem = items[currentIndex + 1] || null;
	}

	async function handleSwitchItem(action: 'next' | 'previous') {
		let gotoItemID = -1;
		if (action === 'next') {
			if (!nextItem) {
				fullItemFilter.page = (fullItemFilter.page ?? 1) + 1;
				const { items } = await listItems(fullItemFilter);
				if (items.length == 0) {
					toast.error('No more items');
					return;
				}
				gotoItemID = items[0].id;
			} else {
				gotoItemID = nextItem.id;
			}
		} else {
			if (!prevItem) {
				if (!fullItemFilter.page) {
					toast.error('No more items');
					return;
				}
				fullItemFilter.page = fullItemFilter.page - 1;
				const { items } = await listItems(fullItemFilter);
				if (items.length == 0) {
					toast.error('No more items');
					return;
				}
				gotoItemID = items.at(-1)!.id;
			} else {
				gotoItemID = prevItem.id;
			}
		}

		goto('/items?id=' + gotoItemID, { invalidateAll: true });
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
		<Separator orientation="vertical" class="h-5" />
		<Separator orientation="vertical" class="h-5" />
		<ItemActionBase
			fn={() => {
				handleSwitchItem('previous');
			}}
			tooltip="Previous item"
			icon={ChevronLeftIcon}
		/>
		<ItemActionBase
			fn={() => {
				handleSwitchItem('next');
			}}
			tooltip="Next item"
			icon={ChevronRightIcon}
		/>
	</div>
</div>
