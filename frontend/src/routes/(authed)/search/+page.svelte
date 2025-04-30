<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { applyFilterToURL, parseURLtoFilter } from '$lib/api/item';
	import ItemList from '$lib/components/ItemList.svelte';
	import PageNavHeader from '$lib/components/PageNavHeader.svelte';
	import { t } from '$lib/i18n';
	import { Search } from 'lucide-svelte';

	let { data } = $props();
	let filterForm = $state(Object.assign({}, parseURLtoFilter(page.url.searchParams)));

	async function handleSearch(e: Event) {
		e.preventDefault();

		const url = page.url;
		applyFilterToURL(url, filterForm);
		console.log(url.toString());
		goto(url, {
			invalidate: ['app:page']
		});
	}
</script>

<svelte:head>
	<title>{t('common.search')}: {filterForm.keyword}</title>
</svelte:head>

<div class="flex flex-col">
	<PageNavHeader title={t('common.search')}></PageNavHeader>
	<div class="px-4 lg:px-8">
		<div class="py-6">
			<h1 class="text-3xl font-bold">{t('common.search')}: {filterForm.keyword}</h1>
		</div>
		<form onsubmit={handleSearch} class="w-full max-w-lg pb-4">
			<div class="join w-full">
				<div class="w-full">
					<label class="input join-item w-full">
						<Search class="size-4 opacity-50" />
						<input
							type="search"
							placeholder={t('item.search.placeholder')}
							bind:value={filterForm.keyword}
							required
						/>
					</label>
				</div>
				<button type="submit" class="btn btn-primary join-item">{t('common.search')}</button>
			</div>
		</form>
		<ItemList data={data.items} highlightUnread={true} />
	</div>
</div>
