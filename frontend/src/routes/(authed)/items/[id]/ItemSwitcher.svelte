<script lang="ts">
	import { hotkey, shortcuts } from '$lib/components/ShortcutHelpModal.svelte';
	import { ChevronLeft, ChevronRight } from 'lucide-svelte';

	interface Props {
		itemID: number;
		itemsQueue: number[];
		action: 'next' | 'previous';
	}
	let { itemID, itemsQueue, action }: Props = $props();

	let currentIndex = $derived(itemsQueue.findIndex((id) => id === itemID));
	let prevID = $derived(itemsQueue[currentIndex - 1] ?? itemsQueue[0]);
	let nextID = $derived(itemsQueue[currentIndex + 1] ?? itemsQueue.at(-1));
	let goto = $derived((action === 'previous' ? prevID : nextID) ?? itemID);
</script>

<a
	href={'/items/' + goto}
	use:hotkey={action === 'previous' ? shortcuts.prevItem.keys : shortcuts.nextItem.keys}
	class={`btn lg:btn-ghost btn-circle lg:btn-xl fixed bottom-1 ${action === 'previous' ? 'left-1' : 'right-1'} lg:sticky lg:top-[50%] ${goto !== itemID ? '' : 'invisible'}`}
>
	{#if action === 'previous'}
		<ChevronLeft />
	{:else}
		<ChevronRight />
	{/if}
</a>
