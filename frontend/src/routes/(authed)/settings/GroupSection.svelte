<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { allGroups, createGroup, deleteGroup, updateGroup } from '$lib/api/group';
	import type { Group } from '$lib/api/model';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import Section from './Section.svelte';

	let newGroup = $state('');
	let existingGroups: Group[] = $state([]);
	onMount(async () => {
		const resp = await allGroups();
		existingGroups = resp;
	});

	async function handleAddNew() {
		try {
			await createGroup(newGroup);
			toast.success('Create successfully');
		} catch (e) {
			toast.error((e as Error).message);
		}
		invalidateAll();
	}

	async function handleUpdate(id: number) {
		const group = existingGroups.find((v) => v.id === id);
		if (!group) return;
		try {
			await updateGroup(id, group.name);
			toast.success('Update successfully');
		} catch (e) {
			toast.error((e as Error).message);
		}
		invalidateAll();
	}

	async function handleDelete(id: number) {
		if (
			!confirm(
				'Are you sure you want to delete this group? All its feeds will be moved to the default group'
			)
		)
			return;
		if (id === 1) {
			toast.error('Cannot delete the default group');
			return;
		}
		try {
			await deleteGroup(id);
			toast.success('Delete successfully');
		} catch (e) {
			toast.error((e as Error).message);
		}
		invalidateAll();
	}
</script>

<Section id="groups" title="Group Management" description="Group's name should be unique.">
	<div class="flex flex-col gap-2">
		{#each existingGroups as g}
			<div class="flex items-center gap-2">
				<input type="text" class="input" bind:value={g.name} />
				<div>
					<button onclick={() => handleUpdate(g.id)} class="btn btn-ghost"> Save </button>
					<button onclick={() => handleDelete(g.id)} class="btn btn-ghost text-red-600">
						Delete
					</button>
				</div>
			</div>
		{/each}
		<div class="flex items-center gap-2">
			<input type="text" class="input" bind:value={newGroup} />
			<div>
				<button onclick={() => handleAddNew()} class="btn btn-ghost"> Add </button>
			</div>
		</div>
	</div>
</Section>
