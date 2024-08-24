<script lang="ts">
	import type { groupFeeds } from './+page';
	import * as Sheet from '$lib/components/ui/sheet';
	import { Button } from '$lib/components/ui/button';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import { toast } from 'svelte-sonner';
	import { invalidateAll } from '$app/navigation';
	import Input from '$lib/components/ui/input/input.svelte';
	import { CheckIcon, PenIcon, PlusIcon, TrashIcon, XIcon } from 'lucide-svelte';
	import { createGroup, deleteGroup, updateGroup } from '$lib/api/group';

	export let groups: groupFeeds[];
	export let open: boolean;

	let editingGroup: groupFeeds = { id: -1, name: '', feeds: [] };
	let showNew = false;
	let newGroup = '';
	let openDelete = false;

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
						on:change={(e) => (editingGroup.name = e.target.value)}
						disabled={editingGroup.id !== group.id}
						required
					/>
					<div class="flex flex-nowrap gap-1">
						{#if editingGroup.id === group.id}
							<Button size="icon" variant="outline" on:click={handleUpdate}>
								<CheckIcon size="15" />
							</Button>
						{:else}
							<Button size="icon" variant="outline" on:click={() => (editingGroup = group)}>
								<PenIcon size="15" />
							</Button>
						{/if}
						<Button
							size="icon"
							variant="outline"
							class="border-destructive text-destructive hover:bg-destructive hover:text-destructive-foreground"
							on:click={() => {
								editingGroup = group;
								openDelete = true;
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
						<Button size="icon" variant="outline" on:click={handleAddNew}>
							<CheckIcon size="15" />
						</Button>
						<Button size="icon" variant="outline" on:click={() => (showNew = false)}
							><XIcon size="15" /></Button
						>
					</div>
				</div>
			{/if}
		</div>
		<Button class="mt-2 w-full" size="icon" on:click={() => (showNew = true)}><PlusIcon /></Button>
	</Sheet.Content>
</Sheet.Root>

<AlertDialog.Root bind:open={openDelete}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Are you absolutely sure?</AlertDialog.Title>
			<AlertDialog.Description>
				<p>
					This action cannot be undone. This will permanently delete <b>{editingGroup.name}</b>.
				</p>
				{#if editingGroup.feeds.length > 0}
					<p>
						Its <b>{editingGroup.feeds.length}</b> feeds will be moved to the default group
						<b>{groups.find((v) => v.id === 1)?.name}</b>.
					</p>
				{/if}
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action on:click={handleDelete}>Continue</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
