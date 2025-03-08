import { listFeeds } from '$lib/api/feed';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async () => {
	const feeds = await listFeeds();
	return {
		feeds
	};
};
