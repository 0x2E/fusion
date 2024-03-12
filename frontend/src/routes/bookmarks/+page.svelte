<script lang="ts">
	import ItemList from '$lib/components/ItemList.svelte';
	import PageHead from '$lib/components/PageHead.svelte';
</script>

<svelte:head>
	<title>Bookmark</title>
</svelte:head>

<PageHead title="Bookmark" />
<ItemList filter={{ bookmark: true }} />
