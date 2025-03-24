<script lang="ts">
	import { listItems } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import ItemActionBookmark from '$lib/components/ItemActionBookmark.svelte';
	import ItemActionGotoFeed from '$lib/components/ItemActionGotoFeed.svelte';
	import ItemActionUnread from '$lib/components/ItemActionUnread.svelte';
	import ItemActionVisitLink from '$lib/components/ItemActionVisitLink.svelte';
	import PageNavHeader from '$lib/components/PageNavHeader.svelte';
	import { render } from '$lib/render-item';
	import { fullItemFilter } from '$lib/state.svelte';
	import { ExternalLink } from 'lucide-svelte';
	import ItemSwitcher from './ItemSwitcher.svelte';

	let { data } = $props();

	let currentIndex = $state(-1);
	function getIndex(id: number) {
		listItems(fullItemFilter).then((resp) => {
			currentIndex = resp.items.findIndex((v) => v.id === id);
		});
	}
	$effect(() => {
		getIndex(data.id);
	});

	let item = $state<Item>(data);
	$effect(() => {
		item = data;
	});

	let safeContent = $derived(render(data.content, data.link));
</script>

<PageNavHeader title={data.title}>
	<ItemActionGotoFeed {item} />
	<ItemActionUnread bind:item />
	<ItemActionBookmark bind:item />
	<ItemActionVisitLink {item} />
</PageNavHeader>

<div class="relative flex w-full grow justify-around px-4 py-6">
	{#if currentIndex > -1}
		<ItemSwitcher itemID={data.id} index={currentIndex} action="previous" />
	{/if}
	<article class="w-full max-w-prose">
		<div class="space-y-2 pb-8">
			<h1 class="text-4xl font-bold">
				<a
					href={data.link}
					target="_blank"
					class="inline-flex items-center gap-2 no-underline hover:underline"
				>
					<span>
						{data.title || data.link}
					</span>
					<ExternalLink class="hidden size-5 md:block" />
				</a>
			</h1>
			<a href={'/feeds/' + data.feed.id} class="text-base-content/60 text-sm hover:underline">
				{data.feed.name} | {new Date(data.pub_date).toLocaleString()}
			</a>
		</div>
		<div class="prose text-wrap break-words">
			{@html safeContent}
		</div>
	</article>
	{#if currentIndex > -1}
		<ItemSwitcher itemID={data.id} index={currentIndex} action="next" />
	{/if}
</div>
