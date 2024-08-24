<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { updateUnread } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { toast } from 'svelte-sonner';
	import ItemActionBase from './ItemActionBase.svelte';
	import { CheckIcon, UndoIcon } from 'lucide-svelte';

	export let data: Item;
	export let buttonClass = '';
	export let iconSize = 18;

	async function toggleUnread(e: Event) {
		e.preventDefault();
		try {
			await updateUnread([data.id], !data.unread);
			invalidateAll();
		} catch (e) {
			toast.error((e as Error).message);
		}
	}
	$: icon = data.unread ? CheckIcon : UndoIcon;
	$: tooltip = data.unread ? 'Mark as Read' : 'mark as Unread';
</script>

<ItemActionBase fn={toggleUnread} {tooltip} {buttonClass} {icon} {iconSize} />
