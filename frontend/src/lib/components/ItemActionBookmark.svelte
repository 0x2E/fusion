<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { page } from '$app/state';
	import { updateBookmark } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { t } from '$lib/i18n';
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
			invalidate('page:' + page.url.pathname);
		} catch (e) {
			toast.error((e as Error).message);
		}
	}
	let Icon = $derived(data.bookmark ? BookmarkXIcon : BookmarkIcon);
	let tooltip = $derived(
		data.bookmark ? t('item.remove_from_bookmark') : t('item.add_to_bookmark')
	);
</script>

<div class="tooltip tooltip-bottom" data-tip={tooltip}>
	<button onclick={toggleBookmark} aria-label={tooltip} class="btn btn-ghost btn-square">
		<Icon class="size-4" />
	</button>
</div>
