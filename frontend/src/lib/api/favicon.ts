export function getFavicon(feedLink: string): string {
	const domain = new URL(feedLink).hostname;
	return 'https://www.google.com/s2/favicons?sz=32&domain=' + domain;
}
