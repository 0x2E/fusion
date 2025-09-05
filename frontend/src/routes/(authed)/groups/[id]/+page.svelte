<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { refreshFeeds } from '$lib/api/feed.js';
	import ItemActionMarkAllasRead from '$lib/components/ItemActionMarkAllasRead.svelte';
	import ItemList from '$lib/components/ItemList.svelte';
	import PageNavHeader from '$lib/components/PageNavHeader.svelte';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import { t } from '$lib/i18n';
	import { Settings2 } from 'lucide-svelte';

	let { data } = $props();

	async function handleRefresh() {
		await refreshFeeds({ all: true });
		await invalidateAll();
	}
</script>

<svelte:head>
	{#await data.group then group}
		<title>{group.name}</title>
	{/await}
</svelte:head>

<PullToRefresh onRefresh={handleRefresh}>
	{#await data.group then group}
		<PageNavHeader showSearch={true}>
			{#await data.items then items}
				<ItemActionMarkAllasRead items={items.items} />
			{/await}
			<div class="tooltip tooltip-bottom" data-tip={t('common.settings')}>
				<a href="/settings#groups" class="btn btn-ghost btn-square">
					<Settings2 class="size-4" />
				</a>
			</div>
		</PageNavHeader>

		<div class="px-4 lg:px-8">
			<div class="items-center py-6">
				<h1 class="text-3xl font-bold">{group.name}</h1>
			</div>
			<ItemList data={data.items} highlightUnread={true} />
		</div>
	{/await}
</PullToRefresh>
