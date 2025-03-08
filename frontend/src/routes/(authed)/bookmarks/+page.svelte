<script lang="ts">
	import ActionSearch from '$lib/components/ActionSearch.svelte';
	import ItemList from '$lib/components/ItemList.svelte';
	import PageHead from '$lib/components/PageHead.svelte';
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
	<PageNavHeader title="Bookmark">
		<ActionSearch />
	</PageNavHeader>
	<PageHead title="Bookmark" />
	<ItemList items={data.items.data} />
</div>
