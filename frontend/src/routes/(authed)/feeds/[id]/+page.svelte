<script lang="ts">
	import FeedActionRefresh from '$lib/components/FeedActionRefresh.svelte';
	import ItemActionMarkAllasRead from '$lib/components/ItemActionMarkAllasRead.svelte';
	import ItemList from '$lib/components/ItemList.svelte';
	import PageNavHeader from '$lib/components/PageNavHeader.svelte';
	import { t } from '$lib/i18n';
	import ActionMenu from './ActionMenu.svelte';

	let { data } = $props();
</script>

<svelte:head>
	{#await data.feed then feed}
		<title>{feed.name}</title>
	{/await}
</svelte:head>

{#await data.feed}
	Loading...
{:then feed}
	<PageNavHeader showSearch={true}>
		{#await data.items then items}
			<ItemActionMarkAllasRead items={items.items} />
		{/await}
		<FeedActionRefresh {feed} />
		<ActionMenu {feed} />
	</PageNavHeader>

	{#if feed.suspended}
		<div role="alert" class="alert alert-warning alert-soft rounded-none">
			<svg
				xmlns="http://www.w3.org/2000/svg"
				class="size-5 shrink-0 stroke-current"
				fill="none"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
				/>
			</svg>
			<p class="text-sm">{t('feed.banner.suspended')}</p>
		</div>
	{:else if feed.failure}
		<div role="alert" class="alert alert-error alert-soft rounded-none">
			<svg
				xmlns="http://www.w3.org/2000/svg"
				class="size-5 shrink-0 stroke-current"
				fill="none"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
				/>
			</svg>
			<p class="text-sm">{t('feed.banner.failed', { error: feed.failure })}</p>
		</div>
	{/if}

	<div class="px-4 lg:px-8">
		<div class="items-center py-6">
			<h1 class="text-3xl font-bold">{feed.name}</h1>
			<p class="text-base-content/60 text-sm">{feed.link}</p>
		</div>
		<ItemList data={data.items} highlightUnread={true} />
	</div>
{/await}
