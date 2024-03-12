import { api } from './api';
import type { Item } from './model';

export type ListFilter = {
	count?: number;
	offset?: number;
	keyword?: string;
	feed_id?: number;
	unread?: boolean;
	bookmark?: boolean;
};

export async function listItems(options?: ListFilter) {
	if (options) {
		// trip undefinded fields: https://github.com/sindresorhus/ky/issues/293
		options = JSON.parse(JSON.stringify(options));
	}
	return api
		.get('items', {
			searchParams: options
		})
		.json<{ total: number; items: Item[] }>();
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
