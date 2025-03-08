import type { PageLoad } from './$types';
import { getItem } from '$lib/api/item';
import { error } from '@sveltejs/kit';

export const prerender = false;

export const load: PageLoad = ({ params }) => {
	// use searchParams instead of params for static build
	const id = parseInt(params.id || '0');
	if (id < 1) {
		error(404, 'wrong id');
	}
	return getItem(id);
};
