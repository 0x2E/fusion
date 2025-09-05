<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { refreshFeeds } from '$lib/api/feed.js';
	import ItemList from '$lib/components/ItemList.svelte';
	import PageNavHeader from '$lib/components/PageNavHeader.svelte';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import { t } from '$lib/i18n';

	let { data } = $props();

	async function handleRefresh() {
		await refreshFeeds({ all: true });
		await invalidateAll();
	}
</script>

<svelte:head>
	<title>{t('common.all')}</title>
</svelte:head>

<PullToRefresh onRefresh={handleRefresh}>
	<div class="flex flex-col">
		<PageNavHeader showSearch={true}></PageNavHeader>
		<div class="px-4 lg:px-8">
			<div class="py-6">
				<h1 class="text-3xl font-bold">{t('common.all')}</h1>
			</div>
			<ItemList data={data.items} highlightUnread={true} />
		</div>
	</div>
</PullToRefresh>
