<script lang="ts">
	import ItemActionMarkAllasRead from '$lib/components/ItemActionMarkAllasRead.svelte';
	import ItemList from '$lib/components/ItemList.svelte';
	import PageNavHeader from '$lib/components/PageNavHeader.svelte';

	let { data } = $props();
</script>

<svelte:head>
	<title>Unread</title>
</svelte:head>

<div class="flex flex-col">
	<PageNavHeader showSearch={true}>
		<ItemActionMarkAllasRead items={data.items.data} />
	</PageNavHeader>
	<div class="px-4 py-6">
		<h1 class="text-3xl font-bold">Unread</h1>
	</div>
	<ItemList items={data.items.data} total={data.items.total} />
</div>
