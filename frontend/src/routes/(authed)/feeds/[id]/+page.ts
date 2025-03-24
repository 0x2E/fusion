import { getFeed } from '$lib/api/feed';
import { listItems, parseURLtoFilter } from '$lib/api/item';
import { fullItemFilter } from '$lib/state.svelte';
import type { PageLoad } from './$types';

export const prerender = false;

export const load: PageLoad = async ({ depends, url, params }) => {
	depends(`page:${url.pathname}`);

	const id = parseInt(params.id);
	const feed = getFeed(id);
	const filter = parseURLtoFilter(url.searchParams, {
		feed_id: id
	});
	Object.assign(fullItemFilter, filter);
	const items = listItems(filter);
	return { feed: feed, items: items };
};
