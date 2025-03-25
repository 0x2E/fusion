<script lang="ts">
	import { ChevronLeft, ChevronRight } from 'lucide-svelte';

	interface Props {
		itemID: number;
		itemsQueue: number[];
		action: 'next' | 'previous';
	}
	let { itemID, itemsQueue, action }: Props = $props();

	let currentIndex = $derived(itemsQueue.findIndex((id) => id === itemID));
	let prevID = $derived(itemsQueue[currentIndex - 1] ?? -1);
	let nextID = $derived(itemsQueue[currentIndex + 1] ?? -1);
	let goto = $derived(action === 'previous' ? prevID : nextID);
</script>

<a
	href={`/items/${action === 'previous' ? prevID : nextID}`}
	class={`btn lg:btn-ghost btn-circle lg:btn-xl fixed bottom-1 ${action === 'previous' ? 'left-1' : 'right-1'} lg:sticky lg:top-[50%] ${goto === -1 ? 'invisible' : ''}`}
>
	{#if action === 'previous'}
		<ChevronLeft />
	{:else}
		<ChevronRight />
	{/if}
</a>
