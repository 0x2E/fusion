import { getFeed } from '$lib/api/feed';
import { listItems } from '$lib/api/item';
import type { PageLoad } from './$types';

export const prerender = false;

export const load: PageLoad = async ({ params }) => {
	const id = parseInt(params.id);
	const feed = getFeed(id);
	const items = listItems({ feed_id: id });
	return { feed: feed, items: items };
};
