import { listFeeds } from '$lib/api/feed';
import { listItems, parseURLtoFilter } from '$lib/api/item';
import { fullItemFilter } from '$lib/state.svelte';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const filter = parseURLtoFilter(url.searchParams, {
		unread: undefined,
		bookmark: undefined
	});
	Object.assign(fullItemFilter, filter);
	const feeds = await listFeeds();
	const items = await listItems(filter);
	return {
		feeds: feeds,
		items: {
			total: items.total,
			data: items.items
		}
	};
};
