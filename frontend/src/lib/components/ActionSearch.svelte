<script lang="ts">
	import { listItems } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { debounce } from '$lib/utils';
	import { Search } from 'lucide-svelte';
	import ItemList from './ItemList.svelte';
	import { t } from '$lib/i18n';

	let modal = $state<HTMLDialogElement>();
	let keyword = $state('');
	let result = $state<{ total: number; items: Item[] } | null>();
	// let isMac = navigator.platform.indexOf('Mac') === 0 || navigator.platform === 'iPhone';

	const handleSearch = debounce(async () => {
		const resp = await listItems(); // TODO filter
		result = resp;
	}, 500);
</script>

<label class="input input-sm lg:w-80">
	<Search class="size-4 opacity-50" />
	<input
		type="search"
		class="input"
		placeholder={t('item.search.placeholder')}
		onclick={() => modal?.showModal()}
	/>
	<!-- <kbd class="kbd kbd-sm">{isMac ? '⌘' : '^'}</kbd>
	<kbd class="kbd kbd-sm">K</kbd> -->
</label>

<dialog id="search" bind:this={modal} class="modal modal-bottom sm:modal-middle">
	<div class="modal-box min-h-80 w-full overflow-x-hidden sm:max-w-4xl">
		<form method="dialog">
			<button class="btn btn-sm btn-circle btn-ghost absolute top-2 right-2">✕</button>
		</form>
		<h3 class="text-lg font-bold">{t('common.search')}</h3>
		<div class="py-4">
			<label class="input w-full">
				<Search class="size-4 opacity-50" />
				<input
					type="search"
					required
					placeholder={t('item.search.placeholder')}
					bind:value={keyword}
					oninput={handleSearch}
					class="w-full"
				/>
			</label>
			{#if result?.total}
				<div class="mt-6">
					<ItemList items={result.items} total={result.total} />
				</div>
			{/if}
		</div>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button>close</button>
	</form>
</dialog>
