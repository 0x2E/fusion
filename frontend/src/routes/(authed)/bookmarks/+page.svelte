<script lang="ts">
	import ItemList from '$lib/components/ItemList.svelte';
	import ItemListPlaceholder from '$lib/components/ItemListPlaceholder.svelte';
	import PageNavHeader from '$lib/components/PageNavHeader.svelte';
	import { t } from '$lib/i18n';
	import type { PageData } from './$types';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();
</script>

<svelte:head>
	<title>{t('common.bookmark')}</title>
</svelte:head>

<div class="flex flex-col">
	<PageNavHeader showSearch={true}></PageNavHeader>
	<div class="px-4 lg:px-8">
		<div class="py-6">
			<h1 class="text-3xl font-bold">{t('common.bookmark')}</h1>
		</div>
		{#await data.items}
			<ItemListPlaceholder />
		{:then items}
			<ItemList items={items.items} total={items.total} />
		{/await}
	</div>
</div>
