/**
 *  RSSHub paths mapped to their respective hostname.
 *  Sorted by hostname first then by RSSHub path.
 */
const rssHubMap = {
	'/papers/category/arxiv': 'arxiv.org',
	'/trendingpapers/papers': 'arxiv.org',
	'/github': 'github.com',
	'/google': 'google.com',
	'/dockerhub': 'hub.docker.com',
	'/imdb': 'imdb.com',
	'/hackernews': 'news.ycombinator.com',
	'/phoronix': 'phoronix.com',
	'/rsshub': 'rsshub.app',
	'/twitch': 'twitch.tv',
	'/youtube': 'youtube.com'
};

export function getFavicon(feedLink: string): string {
	const url = new URL(feedLink);
	let hostname = url.hostname;

	if (hostname.includes('rsshub')) {
		for (const prefix in rssHubMap) {
			if (url.pathname.startsWith(prefix)) {
				hostname = rssHubMap[prefix as keyof typeof rssHubMap];
				break;
			}
		}
	}

	return 'https://www.google.com/s2/favicons?sz=32&domain=' + hostname;
}
