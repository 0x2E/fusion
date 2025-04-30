import { allGroups } from '$lib/api/group';
import { listItems, parseURLtoFilter } from '$lib/api/item';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const prerender = false;

export const load: PageLoad = async ({ url, params, depends }) => {
	depends('app:page');

	const id = parseInt(params.id);
	const group = allGroups().then((groups) => {
		const group = groups.find((g) => g.id === id);
		if (!group) {
			error(404, 'Group not found');
		}
		return group;
	});
	const filter = parseURLtoFilter(url.searchParams, {
		unread: undefined,
		bookmark: undefined,
		feed_id: undefined,
		group_id: id
	});
	const items = listItems(filter);
	return { group, items: items };
};
