import { listItems, parseURLtoFilter } from '$lib/api/item';
import { fullItemFilter } from '$lib/state.svelte';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url, depends }) => {
	depends('page:search');

	const filter = parseURLtoFilter(url.searchParams);
	Object.assign(fullItemFilter, filter);

	return {
		filter,
		items: listItems(filter)
	};
};
