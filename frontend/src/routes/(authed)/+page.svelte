<script lang="ts">
	import ActionSearch from '$lib/components/ActionSearch.svelte';
	import ItemActionMarkAllasRead from '$lib/components/ItemActionMarkAllasRead.svelte';
	import ItemList from '$lib/components/ItemList.svelte';
	import PageHead from '$lib/components/PageHead.svelte';
	import PageNavHeader from '$lib/components/PageNavHeader.svelte';

	let { data } = $props();
</script>

<svelte:head>
	<title>Unread</title>
</svelte:head>

<div class="flex flex-col">
	<PageNavHeader title="Unread">
		<ActionSearch />
		<ItemActionMarkAllasRead items={data.items.data} />
	</PageNavHeader>
	<PageHead title="Unread" />
	<ItemList items={data.items.data} />
</div>
