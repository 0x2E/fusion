import { listFeeds } from '$lib/api/feed';
import { allGroups } from '$lib/api/group';
import type { Feed } from '$lib/api/model';
import type { PageLoad } from './$types';

export type groupFeeds = {
	id: number;
	name: string;
	feeds: Feed[];
};

export const load: PageLoad = async () => {
	const feeds = await listFeeds();
	const groups = await allGroups();
	const data: groupFeeds[] = [];
	for (const g of groups) {
		const gf: groupFeeds = {
			id: g.id,
			name: g.name,
			feeds: feeds.filter((v) => v.group.id === g.id)
		};
		data.push(gf);
	}

	return { groups: data };
};
