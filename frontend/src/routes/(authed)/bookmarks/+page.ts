import { listItems, parseURLtoFilter } from '$lib/api/item';
import { fullItemFilter } from '$lib/state.svelte';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const filter = parseURLtoFilter(url.searchParams, {
		unread: undefined,
		bookmark: true
	});
	Object.assign(fullItemFilter, filter);

	return {
		items: listItems(filter)
	};
};
