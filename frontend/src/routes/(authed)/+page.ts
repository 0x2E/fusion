import { listFeeds } from '$lib/api/feed';
import { listItems, parseURLtoFilter } from '$lib/api/item';
import { fullItemFilter } from '$lib/state.svelte';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const filter = parseURLtoFilter(url.searchParams);
	filter.unread = true;
	filter.bookmark = undefined;
	Object.assign(fullItemFilter, filter);
	const items = await listItems(filter);
	const feeds = await listFeeds({ have_unread: true });
	return {
		feeds: feeds,
		items: {
			total: items.total,
			data: items.items
		}
	};
};
