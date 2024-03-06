<script lang="ts">
	import type { PageData } from './$types';
	import ItemList from '$lib/components/ItemList.svelte';
	import PageHead from '$lib/components/PageHead.svelte';

	export let data: PageData;
</script>

<svelte:head>
	<title>Unread</title>
</svelte:head>

<PageHead title="Unread" />
<ItemList data={data.items} />
