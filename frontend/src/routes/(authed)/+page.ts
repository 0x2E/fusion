import { listItems, parseURLtoFilter } from '$lib/api/item';
import { fullItemFilter } from '$lib/state.svelte';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ depends, url }) => {
	depends(`page:${url.pathname}`);

	const filter = parseURLtoFilter(url.searchParams, {
		unread: true,
		bookmark: undefined,
		feed_id: undefined
	});
	Object.assign(fullItemFilter, filter);
	return {
		items: listItems(filter)
	};
};
