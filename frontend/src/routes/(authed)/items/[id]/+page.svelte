<script lang="ts">
	import type { Item } from '$lib/api/model';
	import ItemActionBookmark from '$lib/components/ItemActionBookmark.svelte';
	import ItemActionGotoFeed from '$lib/components/ItemActionGotoFeed.svelte';
	import ItemActionUnread from '$lib/components/ItemActionUnread.svelte';
	import ItemActionVisitLink from '$lib/components/ItemActionVisitLink.svelte';
	import ItemActionShareLink from '$lib/components/ItemActionShareLink.svelte';
	import PageNavHeader from '$lib/components/PageNavHeader.svelte';
	import { render } from '$lib/render-item';
	import { ExternalLink } from 'lucide-svelte';
	import ItemSwitcher from './ItemSwitcher.svelte';
	import { listItems, type ListFilter } from '$lib/api/item';
	import { afterNavigate } from '$app/navigation';

	let { data } = $props();

	let item = $state<Item>(data);
	$effect(() => {
		item = data;
	});

	let safeContent = $derived(render(data.content, data.link));

	// we prefetch a list of items as the queue for the item switcher.
	// this is a bit hacky, but it's easier to maintain and it should work for most of use cases.
	const queueSize = 100; // 100 is enough and the response size is about 50kb.
	let itemsQueue = $state<number[]>([]);
	afterNavigate(async ({ from }) => {
		const fromPath = from?.url.pathname;
		if (!fromPath) return;

		const filter: ListFilter = { page: 1, page_size: queueSize };
		if (fromPath.startsWith('/feeds/')) {
			const feedMatch = fromPath.match(/\/feeds\/(\d+)/);
			if (feedMatch) {
				filter.feed_id = parseInt(feedMatch[1], 10);
			}
		} else {
			switch (fromPath) {
				case '/all':
					break;
				case '/':
					filter.unread = true;
					break;
				case '/bookmarks':
					filter.bookmark = true;
					break;
				default:
					return;
			}
		}
		const resp = await listItems(filter);
		itemsQueue = resp.items.map((item) => item.id);
	});
</script>

<PageNavHeader title={data.title}>
	<ItemActionGotoFeed {item} />
	<ItemActionUnread bind:item enableShortcut={true} />
	<ItemActionBookmark bind:item enableShortcut={true} />
	<ItemActionVisitLink {item} enableShortcut={true} />
	<ItemActionShareLink {item} />
</PageNavHeader>

<div class="relative flex w-full grow justify-around px-4 py-6">
	<ItemSwitcher itemID={data.id} {itemsQueue} action="previous" />
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
	<ItemSwitcher itemID={data.id} {itemsQueue} action="next" />
</div>
