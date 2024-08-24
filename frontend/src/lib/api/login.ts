import { api } from './api';

export async function login(password: string) {
	return api.post('sessions', {
		json: {
			password: password
		}
	});
}

export async function logout() {
	return api.delete('sessions');
}
