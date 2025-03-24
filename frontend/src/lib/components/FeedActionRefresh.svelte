<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { refreshFeeds } from '$lib/api/feed';
	import type { Feed } from '$lib/api/model';
	import { t } from '$lib/i18n';
	import { RefreshCcw } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		feed?: Feed;
		all?: boolean;
	}

	let { feed, all }: Props = $props();

	async function handleRefresh() {
		if (all) {
			if (!confirm(t('feed.refresh.all.confirm'))) {
				return;
			}
		}
		toast.promise(refreshFeeds({ id: feed?.id, all: all }), {
			success: () => {
				invalidateAll();
				if (all) {
					return t('feed.refresh.all.run_in_background');
				}
				return t('state.success');
			},
			error: (e) => {
				invalidateAll();
				console.log(e);
				return String(e);
			}
		});
	}

	let tooltip = $derived(all ? t('feed.refresh.all') : t('feed.refresh'));
</script>

<div class="tooltip tooltip-bottom" data-tip={tooltip}>
	<button onclick={handleRefresh} class="btn btn-ghost btn-square">
		<RefreshCcw class="size-4" />
	</button>
</div>
