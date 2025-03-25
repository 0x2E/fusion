import { listItems, parseURLtoFilter } from '$lib/api/item';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url, depends }) => {
	depends(`page:${url.pathname}`);

	const filter = parseURLtoFilter(url.searchParams, {
		unread: undefined,
		bookmark: undefined,
		feed_id: undefined
	});
	return {
		items: listItems(filter)
	};
};
