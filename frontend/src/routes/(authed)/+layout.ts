import { listFeeds } from '$lib/api/feed';
import { allGroups } from '$lib/api/group';
import { globalState } from '$lib/state.svelte';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async () => {
	const feeds = listFeeds().then((feeds) => {
		globalState.feeds = feeds;
		return feeds;
	});
	const groups = allGroups().then((groups) => {
		groups.sort((a, b) => a.id - b.id);
		globalState.groups = groups;
		return groups;
	});
	return {
		feeds,
		groups
	};
};
