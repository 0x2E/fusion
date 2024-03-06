<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import ItemList from '$lib/components/ItemList.svelte';
	import PageHead from '$lib/components/PageHead.svelte';
	import { Loader2Icon } from 'lucide-svelte';
	import type { PageData } from './$types';
	import { listItems } from '$lib/api/item';
	import { toast } from 'svelte-sonner';

	export let data: PageData;
	$: items = data.items;
	const limit = 10;
	let page = 1;
	let loading = false;

	async function handleLoadMore() {
		loading = true;
		try {
			const resp = await listItems({ offset: page * limit, count: limit });
			if (resp.items.length === 0) {
				toast.warning('No more items');
			} else {
				items.push(...resp.items);
				items = items;
			}
			page += 1;
		} catch (e) {
			toast.error((e as Error).message);
		}
		loading = false;
	}
</script>

<svelte:head>
	<title>All</title>
</svelte:head>

<PageHead title="All" />
<ItemList data={items} />
{#if items.length > 0}
	<Button variant="secondary" class="w-full mt-6" disabled={loading} on:click={handleLoadMore}>
		{#if !loading}
			Load More
		{:else}
			<Loader2Icon class="mr-2 h-4 w-4 animate-spin" />
		{/if}
	</Button>
{/if}
