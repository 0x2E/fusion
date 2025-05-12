<script lang="ts">
	import type { Item } from '$lib/api/model';
	import { t } from '$lib/i18n';
	import { Share2 } from 'lucide-svelte';
	import { activateShortcut, deactivateShortcut, shortcuts } from './ShortcutHelpModal.svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		item: Item;
		enableShortcut?: boolean;
	}

	let { item, enableShortcut }: Props = $props();

	// only show button if native share is available
	const isShareSupported = !!navigator?.share;

	let el = $state<HTMLElement>();
	$effect(() => {
		if (!el) return;

		if (enableShortcut) {
			activateShortcut(el, shortcuts.shareItem.keys);
		} else {
			deactivateShortcut(el);
		}
	});

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
		<button bind:this={el} class="btn btn-ghost btn-square" onclick={shareItem}>
			<Share2 class="size-4" />
		</button>
	</div>
{/if}
