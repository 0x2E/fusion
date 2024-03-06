import { api } from './api';
import type { Feed } from './model';

export async function allFeeds() {
	const resp = await api.get('feeds').json<{ feeds: Feed[] }>();
	return resp.feeds;
}

export async function getFeed(id: number) {
	return await api.get('feed/' + id).json<Feed>();
}

export async function checkValidity(link: string) {
	const resp = await api
		.post('feeds/validation', {
			json: { link: link }
		})
		.json<{ feed_links: { title: string; link: string }[] }>();
	return resp.feed_links;
}

export async function createFeed(data: {
	group_id: number;
	feeds: { name: string; link: string }[];
}) {
	const feeds = data.feeds.map((v) => {
		return { name: v.name, link: v.link };
	});
	return await api.post('feeds', {
		json: { feeds: feeds, group_id: data.group_id }
	});
}

export async function updateFeed(data: Feed) {
	return await api.patch('feeds/' + data.id, {
		json: { name: data.name, link: data.link, group_id: data.group.id }
	});
}

export async function deleteFeed(id: number) {
	return await api.delete('feeds/' + id);
}

export async function refreshFeeds(options: { id?: number; all?: boolean }) {
	return await api.post('feeds/refresh', {
		json: {
			id: options.id,
			all: options.all
		}
	});
}
