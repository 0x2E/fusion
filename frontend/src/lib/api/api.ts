import ky from 'ky';

export const api = ky.create({
	prefixUrl: '/api',
	retry: 0,
	throwHttpErrors: true,
	credentials: 'same-origin',
	hooks: {
		beforeError: [
			// https://github.com/sindresorhus/ky/issues/412
			async (error) => {
				const { response } = error;
				switch (response.status) {
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
