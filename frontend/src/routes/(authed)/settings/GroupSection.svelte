<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { createGroup, deleteGroup, updateGroup } from '$lib/api/group';
	import { globalState } from '$lib/state.svelte';
	import { toast } from 'svelte-sonner';
	import Section from './Section.svelte';
	import { t } from '$lib/i18n';

	let newGroup = $state('');
	let shuffleArticles = $state(localStorage.getItem("shuffleArticles") === 'true');
	let shuffleSeed = $state(parseInt(localStorage.getItem("shuffleSeed")));
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
	async function handleShuffle(){
		shuffleArticles = !shuffleArticles;
		localStorage.setItem("shuffleArticles" , shuffleArticles);
		localStorage.setItem("shuffleSeed", Math.floor(new Date().getTime() / 1000))
	}
</script>

<Section id="groups" title={t('common.groups')} description={t('settings.groups.description')}>
	<div class="flex flex-col space-y-4">
		<div class="flex flex-col items-center space-x-2 md:flex-row">
		<label class="inline-flex items-center cursor-pointer">
			<input type="checkbox"  class="sr-only peer" onclick={handleShuffle} checked={shuffleArticles}>
			<div class="relative w-11 h-6 bg-base-300 rounded-full peer after:rounded-full after:bg-white after:h-5 after:w-5 after:transition-all peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:border-gray-300 after:border after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600 dark:peer-checked:bg-blue-600"></div>
			<span class="ms-3 text-sm font-medium text">{t('settings.groups.shuffle')}</span>
		</label>
		</div>
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
