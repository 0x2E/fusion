<script lang="ts">
	import ItemList from '$lib/components/ItemList.svelte';
	import PageHead from '$lib/components/PageHead.svelte';
	import type { PageData } from './$types';

	export let data: PageData;
</script>

<svelte:head>
	<title>Unread</title>
</svelte:head>

<PageHead title="Unread" />
<ItemList {data} />
