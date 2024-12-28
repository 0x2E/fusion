<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { updateUnread } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { toast } from 'svelte-sonner';
	import ItemActionBase from './ItemActionBase.svelte';
	import { CheckIcon, UndoIcon } from 'lucide-svelte';

	interface Props {
		data: Item;
		buttonClass?: string;
		iconSize?: number;
	}

	let { data, buttonClass = '', iconSize = 18 }: Props = $props();

	async function toggleUnread(e: Event) {
		e.preventDefault();
		try {
			await updateUnread([data.id], !data.unread);
			invalidateAll();
		} catch (e) {
			toast.error((e as Error).message);
		}
	}
	let icon = $derived(data.unread ? CheckIcon : UndoIcon);
	let tooltip = $derived(data.unread ? 'Mark as Read' : 'mark as Unread');
</script>

<ItemActionBase fn={toggleUnread} {tooltip} {buttonClass} {icon} {iconSize} />
