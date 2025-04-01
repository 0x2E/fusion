import { type Feed, type Group } from './api/model';

export const globalState = $state({
	groups: [] as Group[],
	feeds: [] as Feed[]
});
