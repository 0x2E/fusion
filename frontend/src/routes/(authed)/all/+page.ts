import { listItems, parseURLtoFilter } from '$lib/api/item';
import { fullItemFilter } from '$lib/state.svelte';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const filter = parseURLtoFilter(url.searchParams);
	filter.unread = undefined;
	filter.bookmark = undefined;
	Object.assign(fullItemFilter, filter);

	return {
		items: listItems(filter)
	};
};
