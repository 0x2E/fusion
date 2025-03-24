<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { createGroup, deleteGroup, updateGroup } from '$lib/api/group';
	import { globalState } from '$lib/state.svelte';
	import { toast } from 'svelte-sonner';
	import Section from './Section.svelte';
	import { t } from '$lib/i18n';

	let newGroup = $state('');
	const existingGroups = $derived(globalState.groups);

	async function handleAddNew() {
		try {
			await createGroup(newGroup);
			toast.success(t('state.success'));
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
			toast.success(t('state.success'));
		} catch (e) {
			toast.error((e as Error).message);
		}
		invalidateAll();
	}

	async function handleDelete(id: number) {
		if (!confirm(t('settings.groups.delete.confirm'))) return;
		if (id === 1) {
			toast.error(t('settings.groups.delete.error.delete_the_default'));
			return;
		}
		try {
			await deleteGroup(id);
			toast.success(t('state.success'));
		} catch (e) {
			toast.error((e as Error).message);
		}
		invalidateAll();
	}
</script>

<Section id="groups" title={t('common.groups')} description={t('settings.groups.description')}>
	<div class="flex flex-col space-y-4">
		{#each existingGroups as g}
			<div class="flex flex-col items-center space-x-2 md:flex-row">
				<input type="text" class="input w-full md:w-56" bind:value={g.name} />
				<div class="flex gap-2">
					<button onclick={() => handleUpdate(g.id)} class="btn btn-ghost">
						{t('common.save')}
					</button>
					<button onclick={() => handleDelete(g.id)} class="btn btn-ghost text-error">
						{t('common.delete')}
					</button>
				</div>
			</div>
		{/each}
		<div class="flex items-center space-x-2">
			<input type="text" class="input w-full md:w-56" bind:value={newGroup} />
			<button onclick={() => handleAddNew()} class="btn btn-ghost"> {t('common.add')} </button>
		</div>
	</div>
</Section>
