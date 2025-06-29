<script lang="ts">
	import type { Item } from '$lib/api/model';
	import { t } from '$lib/i18n';
	import { Share2 } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		item: Item;
	}

	let { item }: Props = $props();

	// only show button if native share is available
	const isShareSupported = !!navigator?.share;

	function shareItem() {
		try {
			navigator.share({
				title: item.title,
				url: item.link
			});
		} catch (e) {
			toast.error((e as Error).message);
		}
	}
</script>

{#if isShareSupported}
	<div class="tooltip tooltip-bottom" data-tip={t('item.share')}>
		<button class="btn btn-ghost btn-square" onclick={shareItem}>
			<Share2 class="size-4" />
		</button>
	</div>
{/if}
