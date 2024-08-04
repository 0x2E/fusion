import { listFeeds } from '$lib/api/feed';
import { listItems, parseURLtoFilter } from '$lib/api/item';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const filter = parseURLtoFilter(url.searchParams);
	filter.unread = undefined;
	filter.bookmark = undefined;
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
