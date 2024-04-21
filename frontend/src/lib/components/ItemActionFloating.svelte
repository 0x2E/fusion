<script lang="ts">
	import {
		ArrowUpIcon,
		BookmarkIcon,
		BookmarkXIcon,
		CheckIcon,
		ExternalLinkIcon,
		UndoIcon
	} from 'lucide-svelte';
	import type { ComponentType } from 'svelte';
	import type { Icon } from 'lucide-svelte';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { Button } from './ui/button';
	import { Separator } from './ui/separator';
	import { toast } from 'svelte-sonner';
	import type { Item } from '$lib/api/model';
	import { updateBookmark, updateUnread } from '$lib/api/item';
	import { invalidateAll } from '$app/navigation';

	export let data: Item;
	export let fixed = true;

	function getActions(
		unread: boolean,
		bookmark: boolean
	): { icon: ComponentType<Icon>; tooltip: string; handler: (e: Event) => void }[] {
		const visitOriginalAction = [
			{
				icon: ExternalLinkIcon,
				tooltip: 'Visit Original Link',
				handler: handleExternalLink
			},
			{
				icon: ArrowUpIcon,
				tooltip: 'Back to Top',
				handler: handleScrollTop
			}
		];
		const unreadAction = unread
			? { icon: CheckIcon, tooltip: 'Mark as Read', handler: handleToggleUnread }
			: { icon: UndoIcon, tooltip: 'Mark as Unread', handler: handleToggleUnread };
		const bookmarkAction = bookmark
			? { icon: BookmarkXIcon, tooltip: 'Cancel Bookmark', handler: handleToggleBookmark }
			: { icon: BookmarkIcon, tooltip: 'Add to Bookmark', handler: handleToggleBookmark };

		return [unreadAction, bookmarkAction, ...visitOriginalAction];
	}
	$: actions = getActions(data.unread, data.bookmark);

	async function handleToggleUnread(e: Event) {
		e.preventDefault();
		try {
			await updateUnread([data.id], !data.unread);
			invalidateAll();
		} catch (e) {
			toast.error((e as Error).message);
		}
	}

	async function handleToggleBookmark(e: Event) {
		e.preventDefault();
		try {
			await updateBookmark(data.id, !data.bookmark);
			invalidateAll();
		} catch (e) {
			toast.error((e as Error).message);
		}
	}

	function handleExternalLink(e: Event) {
		e.preventDefault();
		window.open(data.link, '_blank');
	}

	function handleScrollTop(e: Event) {
		e.preventDefault();
		document.body.scrollIntoView({ behavior: 'smooth' });
	}
</script>

<div class="{fixed ? 'fixed' : ''} bottom-2 left-0 right-0">
	<div
		class="flex flex-row justify-center items-center gap-2 rounded-full border shadow w-fit mx-auto bg-background px-6 py-2"
	>
		{#each actions as action, index}
			<Tooltip.Root>
				<Tooltip.Trigger asChild let:builder>
					<Button
						builders={[builder]}
						variant="ghost"
						on:click={action.handler}
						size="icon"
						aria-label={action.tooltip}
						class="rounded-full"
					>
						<svelte:component this={action.icon} size="20" />
					</Button>
				</Tooltip.Trigger>
				<Tooltip.Content>
					<p>{action.tooltip}</p>
				</Tooltip.Content>
			</Tooltip.Root>
			{#if index !== actions.length - 1}
				<Separator orientation="vertical" class="h-5" />
			{/if}
		{/each}
	</div>
</div>
