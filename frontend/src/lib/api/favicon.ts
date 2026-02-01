// RSSHub path prefix to actual domain mapping
const rssHubMap: Record<string, string> = {
  arxiv: "arxiv.org",
  github: "github.com",
  google: "google.com",
  dockerhub: "hub.docker.com",
  imdb: "imdb.com",
  hackernews: "news.ycombinator.com",
  reddit: "reddit.com",
  twitter: "twitter.com",
  youtube: "youtube.com",
  bilibili: "bilibili.com",
  telegram: "telegram.org",
  instagram: "instagram.com",
  linkedin: "linkedin.com",
  medium: "medium.com",
  producthunt: "producthunt.com",
  sspai: "sspai.com",
};

/**
 * Get favicon URL for a feed.
 * Uses site_url if available, otherwise extracts domain from feed link.
 */
export function getFaviconUrl(feedLink: string, siteUrl?: string): string {
  let domain: string;

  if (siteUrl) {
    try {
      domain = new URL(siteUrl).hostname;
    } catch {
      domain = extractDomainFromFeedLink(feedLink);
    }
  } else {
    domain = extractDomainFromFeedLink(feedLink);
  }

  return `https://www.google.com/s2/favicons?domain=${domain}&sz=32`;
}

function extractDomainFromFeedLink(feedLink: string): string {
  try {
    const url = new URL(feedLink);
    const hostname = url.hostname;

    // Handle RSSHub feeds
    if (hostname.includes("rsshub")) {
      const pathPrefix = url.pathname.split("/")[1];
      if (pathPrefix && rssHubMap[pathPrefix]) {
        return rssHubMap[pathPrefix];
      }
    }

    return hostname;
  } catch {
    return "example.com";
  }
}
