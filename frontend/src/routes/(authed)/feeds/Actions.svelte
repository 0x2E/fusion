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

	interface Props {
		groups: groupFeeds[];
	}

	let { groups }: Props = $props();

	let openRefreshAll = $state(false);
	let openAdd = $state(false);
	let openGroups = $state(false);
	let openOPML = $state(false);

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
			<Tooltip.Trigger asChild >
				{#snippet children({ builder })}
								<Button
						builders={[builder]}
						variant="outline"
						on:click={action.handler}
						size="icon"
						aria-label={action.tooltip}
					>
						<action.icon size="20" />
					</Button>
											{/snippet}
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
