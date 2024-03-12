<script lang="ts">
	import ItemList from '$lib/components/ItemList.svelte';
	import PageHead from '$lib/components/PageHead.svelte';
</script>

<svelte:head>
	<title>Unread</title>
</svelte:head>

<PageHead title="Unread" />
<ItemList filter={{ unread: true }} />
