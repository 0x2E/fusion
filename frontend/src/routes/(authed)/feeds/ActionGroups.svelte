<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { createGroup, deleteGroup, updateGroup } from '$lib/api/group';
	import { Button } from '$lib/components/ui/button';
	import Input from '$lib/components/ui/input/input.svelte';
	import * as Sheet from '$lib/components/ui/sheet';
	import { CheckIcon, PenIcon, PlusIcon, TrashIcon, XIcon } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import type { groupFeeds } from './+page';

	interface Props {
		groups: groupFeeds[];
		open: boolean;
	}

	let { groups, open = $bindable() }: Props = $props();

	let editingGroup: groupFeeds = $state({ id: -1, name: '', feeds: [] });
	let showNew = $state(false);
	let newGroup = $state('');

	async function handleAddNew() {
		try {
			await createGroup(newGroup);
			toast.success(newGroup + ' has been created');
		} catch (e) {
			toast.error((e as Error).message);
		}
		newGroup = '';
		invalidateAll();
	}

	async function handleUpdate() {
		try {
			await updateGroup(editingGroup.id, editingGroup.name);
			toast.success(editingGroup.name + ' has been saved');
		} catch (e) {
			toast.error((e as Error).message);
		}
		invalidateAll();
	}

	async function handleDelete() {
		if (editingGroup.id === 1) {
			toast.error('Cannot delete the default group');
			return;
		}
		try {
			await deleteGroup(editingGroup.id);
			toast.success(editingGroup.name + ' has been deleted');
		} catch (e) {
			toast.error((e as Error).message);
		}
		invalidateAll();
	}
</script>

<Sheet.Root bind:open>
	<Sheet.Content class="w-full md:w-auto overflow-scroll">
		<Sheet.Header>
			<Sheet.Title>Manage Groups</Sheet.Title>
			<Sheet.Description>Group's name should be unique.</Sheet.Description>
		</Sheet.Header>
		<div class="grid gap-2">
			{#each groups as group}
				<div class="flex gap-2">
					<Input
						value={group.name}
						onchange={(e) => (editingGroup.name = e.target.value)}
						disabled={editingGroup.id !== group.id}
						required
					/>
					<div class="flex flex-nowrap gap-1">
						{#if editingGroup.id === group.id}
							<Button size="icon" variant="outline" onclick={handleUpdate}>
								<CheckIcon size="15" />
							</Button>
						{:else}
							<Button size="icon" variant="outline" onclick={() => (editingGroup = group)}>
								<PenIcon size="15" />
							</Button>
						{/if}
						<Button
							size="icon"
							variant="outline"
							class="border-destructive text-destructive hover:bg-destructive hover:text-destructive-foreground"
							onclick={() => {
								editingGroup = group;
								if (
									!confirm(
										`Are you sure you want to delete this group? ${group.feeds.length > 0 ? 'All its feeds will be moved to the default group' : ''}`
									)
								) {
									return;
								}
								handleDelete();
							}}
							disabled={group.id === 1}
						>
							<TrashIcon size="15" />
						</Button>
					</div>
				</div>
			{/each}
			{#if showNew}
				<div class="flex gap-2">
					<Input bind:value={newGroup} placeholder="group name" required />
					<div class="flex flex-nowrap gap-1">
						<Button size="icon" variant="outline" onclick={handleAddNew}>
							<CheckIcon size="15" />
						</Button>
						<Button size="icon" variant="outline" onclick={() => (showNew = false)}
							><XIcon size="15" /></Button
						>
					</div>
				</div>
			{/if}
		</div>
		<Button class="mt-2 w-full" size="icon" onclick={() => (showNew = true)}><PlusIcon /></Button>
	</Sheet.Content>
</Sheet.Root>
