import { goto } from '$app/navigation';
import type { HandleClientError } from '@sveltejs/kit';
import { HTTPError } from 'ky';

export const handleError: HandleClientError = async ({ error }) => {
	console.log(error);
	if (error instanceof HTTPError) {
		if (error.response.status === 401) {
			await goto('/login');
		}
		return { message: error.message };
	}
	return error;
};
