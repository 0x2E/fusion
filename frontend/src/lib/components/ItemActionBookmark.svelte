<script module>
	import { updateBookmark } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { toast } from 'svelte-sonner';

	export async function toggleBookmark(item: Item) {
		try {
			await updateBookmark(item.id, !item.bookmark);
			item.bookmark = !item.bookmark;
		} catch (e) {
			toast.error((e as Error).message);
		}
	}
</script>

<script lang="ts">
	import { t } from '$lib/i18n';
	import { BookmarkIcon, BookmarkXIcon } from 'lucide-svelte';
	import { activateHotkey, deactivateHotkey, shortcuts } from './ShortcutHelpModal.svelte';

	let { item = $bindable<Item>(), enableHotkey = false } = $props();

	let Icon = $derived(item.bookmark ? BookmarkXIcon : BookmarkIcon);
	let tooltip = $derived(
		item.bookmark ? t('item.remove_from_bookmark') : t('item.add_to_bookmark')
	);

	let el = $state<HTMLElement>();
	$effect(() => {
		if (!el) return;

		if (enableHotkey) {
			activateHotkey(el, shortcuts.toggleBookmark.keys);
		} else {
			deactivateHotkey(el);
		}
	});

	function handleClick(e: Event) {
		e.preventDefault();

		toggleBookmark(item);
	}
</script>

<div class="tooltip tooltip-bottom" data-tip={tooltip}>
	<button
		onclick={handleClick}
		aria-label={tooltip}
		bind:this={el}
		class="btn btn-ghost btn-square"
	>
		<Icon class="size-4" />
	</button>
</div>
