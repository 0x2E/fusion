<script lang="ts">
	import { refreshFeeds } from '$lib/api/feed';
	import { buttonVariants } from '$lib/components/ui/button';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import {
		FileSpreadsheetIcon,
		FolderTreeIcon,
		PlusCircleIcon,
		RefreshCcwIcon,
		type Icon as IconType
	} from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import type { groupFeeds } from './+page';
	import ActionAdd from './ActionAdd.svelte';
	import ActionGroups from './ActionGroups.svelte';
	import ActionOpml from './ActionOPML.svelte';

	interface Props {
		groups: groupFeeds[];
	}

	let { groups }: Props = $props();

	let openAdd = $state(false);
	let openGroups = $state(false);
	let openOPML = $state(false);

	const actions: { icon: typeof IconType; tooltip: string; handler: () => void }[] = [
		{
			icon: RefreshCcwIcon,
			tooltip: 'Refresh All Feeds',
			handler: async () => {
				if (!confirm('Are you sure you want to refresh all feeds except the suspended ones?')) {
					return;
				}
				try {
					await refreshFeeds({ all: true });
					toast.success('Start refreshing in the background');
				} catch (e) {
					toast.error((e as Error).message);
				}
			}
		},
		{
			icon: PlusCircleIcon,
			tooltip: 'Add Feed',
			handler: () => {
				openAdd = true;
			}
		},
		{
			icon: FolderTreeIcon,
			tooltip: 'Manage Groups',
			handler: () => {
				openGroups = true;
			}
		},
		{
			icon: FileSpreadsheetIcon,
			tooltip: 'Import/Export Feeds',
			handler: () => {
				openOPML = true;
			}
		}
	];
</script>

<div>
	{#each actions as action}
		<Tooltip.Provider>
			<Tooltip.Root delayDuration={100}>
				<Tooltip.Trigger
					onclick={action.handler}
					aria-label={action.tooltip}
					class={buttonVariants({ variant: 'outline', size: 'icon' })}
				>
					<action.icon size="20" />
				</Tooltip.Trigger>
				<Tooltip.Content>
					<p>{action.tooltip}</p>
				</Tooltip.Content>
			</Tooltip.Root>
		</Tooltip.Provider>
	{/each}
</div>

<ActionAdd bind:open={openAdd} {groups} />
<ActionGroups bind:open={openGroups} {groups} />
<ActionOpml bind:open={openOPML} {groups} />
