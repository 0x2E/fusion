import { api } from './api';
import type { Item } from './model';

type listOptions = {
	count?: number;
	offset?: number;
	keyword?: string;
	feed_id?: number;
	unread?: boolean;
};

export async function listItems(options?: listOptions) {
	return api
		.get('items', {
			searchParams: options
		})
		.json<{ total: number; items: Item[] }>();
}

export async function getItem(id: number) {
	return api.get('items/' + id).json<Item>();
}

export async function updateItem(id: number, unread: boolean) {
	return api.patch('items/' + id, {
		json: {
			unread: unread
		}
	});
}
