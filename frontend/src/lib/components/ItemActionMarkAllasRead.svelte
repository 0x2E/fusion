<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { updateUnread } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { CheckCheck } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		items: Item[];
	}

	let { items }: Props = $props();

	async function handleMarkAllAsRead() {
		try {
			const ids = items.map((v) => v.id);
			await updateUnread(ids, false);
			toast.success('Update successfully');
			invalidateAll();
		} catch (e) {
			toast.error((e as Error).message);
		}
	}
</script>

<div class="tooltip tooltip-bottom" data-tip="Mark All as Read">
	<button onclick={handleMarkAllAsRead} class="btn btn-ghost btn-square">
		<CheckCheck class="size-4" />
	</button>
</div>
