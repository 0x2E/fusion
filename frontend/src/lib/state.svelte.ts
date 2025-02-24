import type { ListFilter } from '$lib/api/item';

// this is useful when we need some filter fields that don't
// exist in the URL filter, such as `unread` which is set in
// page load function.
// note that the URL filter is always the source of truth
// and should be used first.
export const fullItemFilter = $state<ListFilter>({});
