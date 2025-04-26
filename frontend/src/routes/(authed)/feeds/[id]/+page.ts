import { getFeed } from '$lib/api/feed';
import { listItems, parseURLtoFilter } from '$lib/api/item';
import type { PageLoad } from './$types';

export const prerender = false;

export const load: PageLoad = async ({ url, params }) => {
	const id = parseInt(params.id);
	const feed = getFeed(id);
	const filter = parseURLtoFilter(url.searchParams, {
		unread: undefined,
		bookmark: undefined,
		feed_id: id
	});
	const items = listItems(filter);
	return { feed: feed, items: items };
};
