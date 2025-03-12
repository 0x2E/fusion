import { listFeeds } from '$lib/api/feed';
import { allGroups } from '$lib/api/group';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async () => {
	const feeds = await listFeeds();
	const groups = await allGroups();
	groups.sort((a, b) => a.id - b.id);
	return {
		feeds,
		groups
	};
};
