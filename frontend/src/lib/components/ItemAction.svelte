<script lang="ts">
	import {
		BookmarkIcon,
		BookmarkXIcon,
		CheckIcon,
		ExternalLinkIcon,
		UndoIcon
	} from 'lucide-svelte';
	import type { ComponentType } from 'svelte';
	import type { Icon } from 'lucide-svelte';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import Button from './ui/button/button.svelte';
	import { updateItem } from '$lib/api/item';
	import { toast } from 'svelte-sonner';
	import type { Item } from '$lib/api/model';

	export let data: Item;

	function getActions(
		unread: boolean,
		bookmark: boolean
	): { icon: ComponentType<Icon>; tooltip: string; handler: (e: Event) => void }[] {
		const visitOriginalAction = {
			icon: ExternalLinkIcon,
			tooltip: 'Visit Original Link',
			handler: handleExternalLink
		};
		const unreadAction = unread
			? { icon: CheckIcon, tooltip: 'Mark as Read', handler: handleToggleUnread }
			: { icon: UndoIcon, tooltip: 'Mark as Unread', handler: handleToggleUnread };
		const bookmarkAction = bookmark
			? { icon: BookmarkXIcon, tooltip: 'Cancel Bookmark', handler: handleToggleBookmark }
			: { icon: BookmarkIcon, tooltip: 'Add to Bookmark', handler: handleToggleBookmark };

		return [unreadAction, bookmarkAction, visitOriginalAction];
	}
	$: actions = getActions(data.unread, data.bookmark);

	// TODO: use invalidateAll after refactoring ItemAction's parents with page load
	async function handleToggleUnread(e: Event) {
		e.preventDefault();
		try {
			await updateItem(data.id, { unread: !data.unread });
			data.unread = !data.unread;
		} catch (e) {
			toast.error((e as Error).message);
		}
	}

	async function handleToggleBookmark(e: Event) {
		e.preventDefault();
		try {
			await updateItem(data.id, { bookmark: !data.bookmark });
			data.bookmark = !data.bookmark;
		} catch (e) {
			toast.error((e as Error).message);
		}
	}

	function handleExternalLink(e: Event) {
		e.preventDefault();
		handleToggleUnread(e);
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
