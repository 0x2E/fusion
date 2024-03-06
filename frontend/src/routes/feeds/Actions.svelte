<script lang="ts">
	import {
		FileSpreadsheetIcon,
		FolderTreeIcon,
		PlusCircleIcon,
		RefreshCcwIcon
	} from 'lucide-svelte';
	import type { ComponentType } from 'svelte';
	import type { Icon } from 'lucide-svelte';
	import type { groupFeeds } from './+page';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { Button } from '$lib/components/ui/button';
	import ActionOpml from './ActionOPML.svelte';
	import ActionAdd from './ActionAdd.svelte';
	import ActionRefreshAll from './ActionRefreshAll.svelte';
	import ActionGroups from './ActionGroups.svelte';

	export let groups: groupFeeds[];

	let openRefreshAll = false;
	let openAdd = false;
	let openGroups = false;
	let openOPML = false;

	const actions: { icon: ComponentType<Icon>; tooltip: string; handler: () => void }[] = [
		{
			icon: RefreshCcwIcon,
			tooltip: 'Refresh All Feeds',
			handler: () => {
				openRefreshAll = true;
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
		<Tooltip.Root>
			<Tooltip.Trigger asChild let:builder>
				<Button builders={[builder]} variant="ghost" on:click={action.handler} size="icon">
					<svelte:component this={action.icon} size="20" />
				</Button>
			</Tooltip.Trigger>
			<Tooltip.Content>
				<p>{action.tooltip}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	{/each}
</div>

<ActionRefreshAll bind:open={openRefreshAll} />
<ActionAdd bind:open={openAdd} {groups} />
<ActionGroups bind:open={openGroups} {groups} />
<ActionOpml bind:open={openOPML} {groups} />
