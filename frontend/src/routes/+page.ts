import { listItems } from '$lib/api/item';
import type { PageLoad } from './$types';

export const load: PageLoad = () => {
	return listItems({ unread: true });
};
