export function parse(content: string) {
	const feeds: { name: string; link: string }[] = [];
	const xmlDoc = new DOMParser().parseFromString(content, 'text/xml');
	const outlines = xmlDoc.getElementsByTagName('outline');

	for (let i = 0; i < outlines.length; i++) {
		const outline = outlines.item(i);
		if (!outline) continue;
		const link = outline.getAttribute('xmlUrl') || outline.getAttribute('htmlUrl') || '';
		if (!link) continue;
		const name = outline.getAttribute('title') || outline.getAttribute('text') || '';
		feeds.push({ name, link });
	}

	return feeds;
}

export function dump(data: { name: string; feeds: { name: string; link: string }[] }[]) {
	const doc = document.implementation.createDocument('', '', null);

	const opmlElement = doc.createElement('opml');
	opmlElement.setAttribute('version', '1.0');

	const headElement = doc.createElement('head');
	const titleElement = doc.createElement('title');
	titleElement.appendChild(doc.createTextNode('Feeds exported from Fusion'));
	headElement.appendChild(titleElement);
	opmlElement.appendChild(headElement);

	const bodyElement = doc.createElement('body');
	for (const group of data) {
		const groupOutlineElement = doc.createElement('outline');
		groupOutlineElement.setAttribute('text', group.name);
		groupOutlineElement.setAttribute('title', group.name);
		for (const feed of group.feeds) {
			const outlineElement = doc.createElement('outline');
			outlineElement.setAttribute('type', 'rss');
			outlineElement.setAttribute('text', feed.name);
			outlineElement.setAttribute('title', feed.name);
			outlineElement.setAttribute('xmlUrl', feed.link);
			outlineElement.setAttribute('htmlUrl', feed.link);
			groupOutlineElement.appendChild(outlineElement);
		}
		bodyElement.appendChild(groupOutlineElement);
	}
	opmlElement.appendChild(bodyElement);

	doc.appendChild(opmlElement);
	return `<?xml version="1.0" encoding="UTF-8"?>` + new XMLSerializer().serializeToString(doc);
}
