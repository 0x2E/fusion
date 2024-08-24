<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { updateBookmark } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { toast } from 'svelte-sonner';
	import ItemActionBase from './ItemActionBase.svelte';
	import { BookmarkXIcon, BookmarkIcon } from 'lucide-svelte';

	export let data: Item;
	export let buttonClass = '';
	export let iconSize = 18;

	async function toggleBookmark(e: Event) {
		e.preventDefault();
		try {
			await updateBookmark(data.id, !data.bookmark);
			invalidateAll();
		} catch (e) {
			toast.error((e as Error).message);
		}
	}
	$: icon = data.bookmark ? BookmarkXIcon : BookmarkIcon;
	$: tooltip = data.bookmark ? 'Cancel Bookmark' : 'Add to Bookmark';
</script>

<ItemActionBase fn={toggleBookmark} {tooltip} {buttonClass} {icon} {iconSize} />
