import { listFeeds } from '$lib/api/feed';
import { allGroups } from '$lib/api/group';
import { setGlobalFeeds, setGlobalGroups } from '$lib/state.svelte';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async ({ depends }) => {
	depends('app:feeds', 'app:groups');

	await Promise.all([
		allGroups().then((groups) => {
			groups.sort((a, b) => a.id - b.id);
			setGlobalGroups(groups);
		}),
		listFeeds().then((feeds) => {
			setGlobalFeeds(feeds);
		})
	]);

	return {};
};
