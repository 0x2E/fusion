<script module>
	import type { Item } from '$lib/api/model';
	import { updateUnread } from '$lib/api/item';
	import { toast } from 'svelte-sonner';

	export async function toggleUnread(item: Item) {
		try {
			await updateUnread([item.id], !item.unread);
			item.unread = !item.unread;
		} catch (e) {
			toast.error((e as Error).message);
		}
	}
</script>

<script lang="ts">
	import { t } from '$lib/i18n';
	import { CheckIcon, UndoIcon } from 'lucide-svelte';
	import { activateHotkey, deactivateHotkey, shortcuts } from './ShortcutHelpModal.svelte';

	let { item = $bindable<Item>(), enableHotkey = false } = $props();

	let Icon = $derived(item.unread ? CheckIcon : UndoIcon);
	let tooltip = $derived(item.unread ? t('item.mark_as_read') : t('item.mark_as_unread'));

	let el = $state<HTMLElement>();
	$effect(() => {
		if (!el) return;

		if (enableHotkey) {
			activateHotkey(el, shortcuts.toggleUnread.keys);
		} else {
			deactivateHotkey(el);
		}
	});

	function handleClick(e: Event) {
		e.preventDefault();
		toggleUnread(item);
	}
</script>

<div class="tooltip tooltip-bottom" data-tip={tooltip}>
	<button onclick={handleClick} bind:this={el} class="btn btn-ghost btn-square">
		<Icon class="size-4" />
	</button>
</div>
