import { goto } from '$app/navigation';
import ky from 'ky';

export const api = ky.create({
	prefixUrl: '/api',
	retry: 0,
	throwHttpErrors: true,
	credentials: 'same-origin',
	hooks: {
		beforeError: [
			async (error) => {
				const { response } = error;
				switch (response.status) {
					case 401:
						await goto('/login');
						break;
					default:
						try {
							const data = await response.json();
							error.message = data.message;
						} catch (e) {
							console.log(e);
						}
				}
				return error;
			}
		]
	}
});
