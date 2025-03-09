<script lang="ts">
	import ItemList from '$lib/components/ItemList.svelte';
	import PageNavHeader from '$lib/components/PageNavHeader.svelte';
	import type { PageData } from './$types';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();
</script>

<svelte:head>
	<title>Bookmark</title>
</svelte:head>

<div class="flex flex-col">
	<PageNavHeader showSearch={true}></PageNavHeader>
	<div class="px-4 py-6">
		<h1 class="text-3xl font-bold">Bookmark</h1>
	</div>
	<ItemList items={data.items.data} total={data.items.total} />
</div>
