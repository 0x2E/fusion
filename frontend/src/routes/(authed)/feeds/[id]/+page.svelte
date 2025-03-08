<script lang="ts">
	import ActionSearch from '$lib/components/ActionSearch.svelte';
	import ItemList from '$lib/components/ItemList.svelte';
	import PageHead from '$lib/components/PageHead.svelte';
	import PageNavHeader from '$lib/components/PageNavHeader.svelte';

	let { data } = $props();
</script>

<svelte:head>
	<title>Feeds</title>
</svelte:head>

{#await data.feed}
	Loading...
{:then feed}
	<PageNavHeader title={feed.name}>
		<ActionSearch />
	</PageNavHeader>
	<PageHead title={feed.name}></PageHead>
	{#await data.items}
		Loading...
	{:then items}
		<ItemList items={items.items} highlightUnread={true} />
	{/await}
{/await}
