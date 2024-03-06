<script lang="ts">
	import { CheckIcon, ExternalLinkIcon, UndoIcon } from 'lucide-svelte';
	import type { ComponentType } from 'svelte';
	import type { Icon } from 'lucide-svelte';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import Button from './ui/button/button.svelte';
	import { updateItem } from '$lib/api/item';
	import { toast } from 'svelte-sonner';
	import { invalidateAll } from '$app/navigation';

	export let data: {
		id: number;
		link: string;
		unread: boolean;
	};

	function getActions(
		unread: boolean
	): { icon: ComponentType<Icon>; tooltip: string; handler: (e: Event) => void }[] {
		const list = [
			// { icon: BookmarkIcon, tooltip: 'Save to Bookmark', handler: handleSaveToBookmark },
			{ icon: ExternalLinkIcon, tooltip: 'Visit Original Link', handler: handleExternalLink }
		];
		const unreadAction = unread
			? { icon: CheckIcon, tooltip: 'Mark as Read', handler: handleMarkAsRead }
			: { icon: UndoIcon, tooltip: 'Mark as Unread', handler: handleMarkAsUnread };
		list.unshift(unreadAction);
		return list;
	}
	$: actions = getActions(data.unread);

	async function handleMarkAsRead(e: Event) {
		e.preventDefault();
		try {
			await updateItem(data.id, false);
		} catch (e) {
			toast.error((e as Error).message);
		}
		invalidateAll();
	}

	async function handleMarkAsUnread(e: Event) {
		e.preventDefault();
		try {
			await updateItem(data.id, true);
		} catch (e) {
			toast.error((e as Error).message);
		}
		invalidateAll();
	}

	function handleExternalLink(e: Event) {
		e.preventDefault();
		handleMarkAsRead(e);
		window.open(data.link, '_target');
	}
</script>

<div>
	{#each actions as action}
		<Tooltip.Root>
			<Tooltip.Trigger asChild let:builder>
				<Button
					builders={[builder]}
					variant="ghost"
					on:click={action.handler}
					class="hover:bg-gray-300 dark:hover:bg-gray-700"
					size="icon"
				>
					<svelte:component this={action.icon} size="18" />
				</Button>
			</Tooltip.Trigger>
			<Tooltip.Content>
				<p>{action.tooltip}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	{/each}
</div>
