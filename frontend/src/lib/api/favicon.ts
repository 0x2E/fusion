/**
 *  RSSHub paths mapped to their respective hostname.
 *  Sorted by hostname first then by pathname.
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

/**
 *  Buttondown.com paths mapped to their respective hostname.
 *  Sorted by hostname first then by pathname.
 */
const buttondownMap = {
	'/denonews/': 'deno.news'
};

export function getFavicon(feedLink: string): string {
	const url = new URL(feedLink);
	let hostname = url.hostname;
	let pathname = url.pathname;

	if (hostname.includes('rsshub')) {
		for (const prefix in rssHubMap) {
			if (pathname.startsWith(prefix)) {
				hostname = rssHubMap[prefix as keyof typeof rssHubMap];
				break;
			}
		}
	}

	if (hostname === 'buttondown.com') {
		for (const prefix in buttondownMap) {
			if (pathname.startsWith(prefix)) {
				hostname = buttondownMap[prefix as keyof typeof buttondownMap];
				break;
			}
		}
	}

	return 'https://www.google.com/s2/favicons?sz=32&domain=' + hostname;
}
