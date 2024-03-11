<script lang="ts">
	import type { PageData } from './$types';
	import ItemList from '$lib/components/ItemList.svelte';
	import PageHead from '$lib/components/PageHead.svelte';

	export let data: PageData;
</script>

<svelte:head>
	<title>Bookmark</title>
</svelte:head>

<PageHead title="Bookmark" />
<ItemList data={data.items} />
