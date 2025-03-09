<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { deleteFeed, updateFeed } from '$lib/api/feed';
	import type { Feed } from '$lib/api/model';
	import { Ellipsis, Pause, Trash } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		feed: Feed;
	}

	let { feed }: Props = $props();

	async function handleToggleSuspended() {
		try {
			await updateFeed(feed.id, {
				suspended: !feed.suspended
			});
			toast.success('Update successfully');
		} catch (e) {
			toast.error((e as Error).message);
		}
		invalidateAll();
	}

	async function handleDelete() {
		if (!confirm(`Are you sure you want to delete [${feed.name}]?`)) return;

		try {
			await deleteFeed(feed.id);
			toast.success('Feed has been deleted');
		} catch (e) {
			toast.error((e as Error).message);
		}

		invalidateAll();
	}
</script>

<div class="dropdown dropdown-end">
	<div tabindex="0" role="button" class="btn btn-ghost btn-square">
		<Ellipsis class="size-4" />
	</div>
	<ul tabindex="0" class="dropdown-content menu bg-base-100 rounded-box z-1 w-52 p-2 shadow-sm">
		<li>
			<button onclick={handleToggleSuspended}>
				<Pause class="size-4" />
				<span>
					{feed.suspended ? 'Resume refreshing' : 'Suspend refreshing'}
				</span>
			</button>
		</li>
		<li>
			<button onclick={handleDelete}>
				<Trash class="size-4" />
				<span> Delete feed</span>
			</button>
		</li>
	</ul>
</div>
