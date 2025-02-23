import { api } from './api';
import type { Item } from './model';

export type ListFilter = {
	page: number;
	page_size: number;
	keyword?: string;
	feed_id?: number;
	unread?: boolean;
	bookmark?: boolean;
};

export async function listItems(options?: ListFilter) {
	if (options) {
		// trip undefinded fields: https://github.com/sindresorhus/ky/issues/293
		options = JSON.parse(JSON.stringify(options));
		sessionStorage.setItem("filter", JSON.stringify(options));
	}
	return await api
		.get('items', {
			searchParams: options
		})
		.json<{ total: number; items: Item[] }>();
}

export function parseURLtoFilter(params: URLSearchParams) {
	const filter: ListFilter = {
		page: parseInt(params.get('page') || '1'),
		page_size: parseInt(params.get('page_size') || '10')
	};
	const keyword = params.get('keyword');
	if (keyword) filter.keyword = keyword;
	const feed_id = params.get('feed_id');
	if (feed_id) filter.feed_id = parseInt(feed_id);
	const unread = params.get('unread');
	if (unread) filter.unread = unread === 'true';
	const bookmark = params.get('bookmark');
	if (bookmark) filter.bookmark = bookmark === 'true';
	return filter;
}

export async function getItem(id: number) {
	return api.get('items/' + id).json<Item>();
}

export async function updateUnread(ids: number[], unread: boolean) {
	return api.patch('items/-/unread', {
		json: {
			ids: ids,
			unread: unread
		}
	});
}

export async function updateBookmark(id: number, bookmark: boolean) {
	return api.patch('items/' + id + '/bookmark', {
		json: {
			bookmark: bookmark
		}
	});
}
