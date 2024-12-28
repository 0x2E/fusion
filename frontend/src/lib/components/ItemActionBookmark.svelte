<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { updateBookmark } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { toast } from 'svelte-sonner';
	import ItemActionBase from './ItemActionBase.svelte';
	import { BookmarkXIcon, BookmarkIcon } from 'lucide-svelte';

	interface Props {
		data: Item;
		buttonClass?: string;
		iconSize?: number;
	}

	let { data, buttonClass = '', iconSize = 18 }: Props = $props();

	async function toggleBookmark(e: Event) {
		e.preventDefault();
		try {
			await updateBookmark(data.id, !data.bookmark);
			invalidateAll();
		} catch (e) {
			toast.error((e as Error).message);
		}
	}
	let icon = $derived(data.bookmark ? BookmarkXIcon : BookmarkIcon);
	let tooltip = $derived(data.bookmark ? 'Cancel Bookmark' : 'Add to Bookmark');
</script>

<ItemActionBase fn={toggleBookmark} {tooltip} {buttonClass} {icon} {iconSize} />
