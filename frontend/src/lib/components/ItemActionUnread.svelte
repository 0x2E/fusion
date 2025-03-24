<script lang="ts">
	import { updateUnread } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { t } from '$lib/i18n';
	import { CheckIcon, UndoIcon } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	let { item = $bindable<Item>() } = $props();

	async function toggleUnread(e: Event) {
		e.preventDefault();
		try {
			await updateUnread([item.id], !item.unread);
			item.unread = !item.unread;
		} catch (e) {
			toast.error((e as Error).message);
		}
	}

	let Icon = $derived(item.unread ? CheckIcon : UndoIcon);
	let tooltip = $derived(item.unread ? t('item.mark_as_read') : t('item.mark_as_unread'));
</script>

<div class="tooltip tooltip-bottom" data-tip={tooltip}>
	<button onclick={toggleUnread} class="btn btn-ghost btn-square">
		<Icon class="size-4" />
	</button>
</div>
