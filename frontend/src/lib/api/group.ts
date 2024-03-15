import { api } from './api';
import type { Group } from './model';

export async function allGroups() {
	const resp = await api.get('groups').json<{ groups: Group[] }>();
	return resp.groups.sort((a, b) => a.id - b.id);
}

export async function createGroup(name: string) {
	return await api.post('groups', {
		json: {
			name: name
		}
	});
}

export async function updateGroup(id: number, name: string) {
	return await api.patch('groups/' + id, {
		json: {
			name: name
		}
	});
}

export async function deleteGroup(id: number) {
	return await api.delete('groups/' + id);
}
