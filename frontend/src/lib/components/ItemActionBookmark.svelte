<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { updateBookmark } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { BookmarkIcon, BookmarkXIcon } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		data: Item;
	}

	let { data }: Props = $props();

	async function toggleBookmark(e: Event) {
		e.preventDefault();
		try {
			await updateBookmark(data.id, !data.bookmark);
			invalidateAll();
		} catch (e) {
			toast.error((e as Error).message);
		}
	}
	let Icon = $derived(data.bookmark ? BookmarkXIcon : BookmarkIcon);
	let tooltip = $derived(data.bookmark ? 'Cancel Bookmark' : 'Add to Bookmark');
</script>

<div class="tooltip tooltip-bottom" data-tip={tooltip}>
	<button onclick={toggleBookmark} aria-label={tooltip} class="btn btn-ghost btn-square">
		<Icon class="size-4" />
	</button>
</div>
