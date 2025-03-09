<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { refreshFeeds } from '$lib/api/feed';
	import type { Feed } from '$lib/api/model';
	import { RefreshCcw } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		feed?: Feed;
		all?: boolean;
	}

	let { feed, all }: Props = $props();

	async function handleRefresh() {
		if (all) {
			if (!confirm('Are you sure you want to refresh all feeds except the suspended ones?')) {
				return;
			}
		}
		toast.promise(refreshFeeds({ id: feed?.id, all: all }), {
			success: () => {
				invalidateAll();
				if (all) {
					return 'Start refreshing in the background';
				}
				return 'Refresh successfully';
			},
			error: (e) => {
				invalidateAll();
				console.log(e);
				return String(e);
			}
		});
	}

	let tooltip = $derived(all ? 'Refresh Feeds' : 'Refresh Feed');
</script>

<div class="tooltip tooltip-bottom" data-tip={tooltip}>
	<button onclick={handleRefresh} class="btn btn-ghost btn-square">
		<RefreshCcw class="size-4" />
	</button>
</div>
