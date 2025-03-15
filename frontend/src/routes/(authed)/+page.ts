import { listFeeds } from '$lib/api/feed';
import { listItems, parseURLtoFilter } from '$lib/api/item';
import { fullItemFilter } from '$lib/state.svelte';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const filter = parseURLtoFilter(url.searchParams, {
		unread: true,
		bookmark: undefined
	});
	Object.assign(fullItemFilter, filter);
	const feeds = await listFeeds({ have_unread: true });
	const items = await listItems(filter);
	return {
		feeds: feeds,
		items: {
			total: items.total,
			data: items.items
		}
	};
};
