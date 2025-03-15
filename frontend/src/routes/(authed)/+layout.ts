import { listFeeds } from '$lib/api/feed';
import { allGroups } from '$lib/api/group';
import { listItems } from '$lib/api/item';
import { fullItemFilter } from '$lib/state.svelte';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async () => {
	const feeds = await listFeeds();
	const groups = await allGroups();
	groups.sort((a, b) => a.id - b.id);

	const filter = { unread: true };
	Object.assign(fullItemFilter, filter);
	const unreadItems = await listItems(filter);

	return {
		feeds,
		groups,
		unreadCount: unreadItems.total
	};
};
