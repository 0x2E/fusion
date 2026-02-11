import type { Group, Feed } from "./api/types";

export interface ParsedFeed {
  name: string;
  link: string;
  siteUrl?: string;
  groupName?: string;
}

/**
 * Parses OPML content and extracts feed information.
 * Supports both flat and nested (outline with children) OPML structures.
 */
export function parseOPML(content: string): ParsedFeed[] {
  const parser = new DOMParser();
  const doc = parser.parseFromString(content, "text/xml");

  const parseError = doc.querySelector("parsererror");
  if (parseError) {
    throw new Error("Invalid OPML format");
  }

  const feeds: ParsedFeed[] = [];
  const body = doc.querySelector("body");
  if (!body) return feeds;

  const processOutline = (outline: Element, groupName?: string) => {
    const xmlUrl = outline.getAttribute("xmlUrl");
    const htmlUrl = outline.getAttribute("htmlUrl");
    const title =
      outline.getAttribute("title") || outline.getAttribute("text") || "";

    if (xmlUrl) {
      // This is a feed
      feeds.push({
        name: title || xmlUrl,
        link: xmlUrl,
        siteUrl: htmlUrl?.trim() || undefined,
        groupName,
      });
    } else {
      // This might be a group/category
      const childOutlines = outline.querySelectorAll(":scope > outline");
      const newGroupName = title || groupName;
      childOutlines.forEach((child) => processOutline(child, newGroupName));
    }
  };

  const topOutlines = body.querySelectorAll(":scope > outline");
  topOutlines.forEach((outline) => processOutline(outline));

  return feeds;
}

/**
 * Generates OPML content from groups and feeds.
 */
export function generateOPML(groups: Group[], feeds: Feed[]): string {
  const feedsByGroup = new Map<number, Feed[]>();
  const ungroupedFeeds: Feed[] = [];

  feeds.forEach((feed) => {
    const group = groups.find((g) => g.id === feed.group_id);
    if (group) {
      const existing = feedsByGroup.get(group.id) || [];
      existing.push(feed);
      feedsByGroup.set(group.id, existing);
    } else {
      ungroupedFeeds.push(feed);
    }
  });

  const escapeXml = (str: string): string =>
    str
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&apos;");

  const feedToOutline = (feed: Feed): string =>
    `      <outline type="rss" text="${escapeXml(feed.name)}" title="${escapeXml(feed.name)}" xmlUrl="${escapeXml(feed.link)}"${feed.site_url ? ` htmlUrl="${escapeXml(feed.site_url)}"` : ""} />`;

  const lines: string[] = [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<opml version="2.0">',
    "  <head>",
    "    <title>Fusion Subscriptions</title>",
    `    <dateCreated>${new Date().toUTCString()}</dateCreated>`,
    "  </head>",
    "  <body>",
  ];

  // Add grouped feeds
  groups.forEach((group) => {
    const groupFeeds = feedsByGroup.get(group.id) || [];
    if (groupFeeds.length > 0) {
      lines.push(
        `    <outline text="${escapeXml(group.name)}" title="${escapeXml(group.name)}">`,
      );
      groupFeeds.forEach((feed) => lines.push(feedToOutline(feed)));
      lines.push("    </outline>");
    }
  });

  // Add ungrouped feeds at top level
  ungroupedFeeds.forEach((feed) => {
    lines.push(
      `    <outline type="rss" text="${escapeXml(feed.name)}" title="${escapeXml(feed.name)}" xmlUrl="${escapeXml(feed.link)}"${feed.site_url ? ` htmlUrl="${escapeXml(feed.site_url)}"` : ""} />`,
    );
  });

  lines.push("  </body>");
  lines.push("</opml>");

  return lines.join("\n");
}

/**
 * Triggers a file download in the browser.
 */
export function downloadFile(
  content: string,
  filename: string,
  mimeType: string,
): void {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}
