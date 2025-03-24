<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { page } from '$app/state';
	import { updateUnread } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { t } from '$lib/i18n';
	import { CheckIcon, UndoIcon } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		data: Item;
	}

	let { data }: Props = $props();

	async function toggleUnread(e: Event) {
		e.preventDefault();
		try {
			await updateUnread([data.id], !data.unread);
			invalidate('page:' + page.url.pathname);
		} catch (e) {
			toast.error((e as Error).message);
		}
	}
	let Icon = $derived(data.unread ? CheckIcon : UndoIcon);
	let tooltip = $derived(data.unread ? t('item.mark_as_read') : t('item.mark_as_unread'));
</script>

<div class="tooltip tooltip-bottom" data-tip={tooltip}>
	<button onclick={toggleUnread} class="btn btn-ghost btn-square">
		<Icon class="size-4" />
	</button>
</div>
