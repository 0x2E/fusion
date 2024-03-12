<script lang="ts">
	import ItemList from '$lib/components/ItemList.svelte';
	import PageHead from '$lib/components/PageHead.svelte';
</script>

<svelte:head>
	<title>All</title>
</svelte:head>

<PageHead title="All" />
<ItemList filter={{}} />
