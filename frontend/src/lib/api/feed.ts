import { api } from './api';
import type { Feed } from './model';

export type FeedListFiler = {
	have_unread?: boolean;
	have_bookmark?: boolean;
};

export async function listFeeds(filter?: FeedListFiler) {
	if (filter) {
		filter = JSON.parse(JSON.stringify(filter));
	}

	const resp = await api
		.get('feeds', {
			searchParams: filter
		})
		.json<{ feeds: Feed[] }>();
	return resp.feeds;
}

export async function getFeed(id: number) {
	return await api.get('feeds/' + id).json<Feed>();
}

export async function checkValidity(link: string) {
	const resp = await api
		.post('feeds/validation', {
			timeout: 20000,
			json: { link: link }
		})
		.json<{ feed_links: { title: string; link: string }[] }>();
	return resp.feed_links;
}

export type FeedCreateForm = {
	group_id: number;
	feeds: {
		name: string;
		link: string;
	}[];
};

export async function createFeed(data: FeedCreateForm) {
	const feeds = data.feeds.map((v) => {
		return { name: v.name, link: v.link };
	});
	return await api.post('feeds', {
		timeout: 20000,
		json: { feeds: feeds, group_id: data.group_id }
	});
}

export type FeedUpdateForm = {
	name?: string;
	link?: string;
	suspended?: boolean;
	req_proxy?: string;
	group_id?: number;
};

export async function updateFeed(id: number, data: FeedUpdateForm) {
	return await api.patch('feeds/' + id, {
		json: data
	});
}

export async function deleteFeed(id: number) {
	return await api.delete('feeds/' + id);
}

export async function refreshFeeds(options: { id?: number; all?: boolean }) {
	return await api.post('feeds/refresh', {
		timeout: 20000,
		json: {
			id: options.id,
			all: options.all
		}
	});
}
