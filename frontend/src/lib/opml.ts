export function parse(content: string) {
	type feedT = {
		name: string;
		link: string;
	};
	type groupT = {
		name: string;
		feeds: feedT[];
	};
	const groups = new Map<string, groupT>();
	const defaultGroup = { name: 'Default', feeds: [] };
	groups.set('Default', defaultGroup);

	function dfs(parentGroup: groupT | null, node: Element) {
		if (node.tagName !== 'outline') {
			return;
		}
		if (node.getAttribute('type')?.toLowerCase() == 'rss') {
			if (!parentGroup) {
				parentGroup = defaultGroup;
			}
			parentGroup.feeds.push({
				name: node.getAttribute('title') || node.getAttribute('text') || '',
				link: node.getAttribute('xmlUrl') || node.getAttribute('htmlUrl') || ''
			});
			return;
		}
		if (!node.children.length) {
			return;
		}
		const nodeName = node.getAttribute('text') || node.getAttribute('title') || '';
		const name = parentGroup ? parentGroup.name + '/' + nodeName : nodeName;
		let curGroup = groups.get(name);
		if (!curGroup) {
			curGroup = { name: name, feeds: [] };
			groups.set(name, curGroup);
		}
		for (const n of node.children) {
			dfs(curGroup, n);
		}
	}

	const xmlDoc = new DOMParser().parseFromString(content, 'text/xml');
	const body = xmlDoc.getElementsByTagName('body')[0];
	if (!body) {
		return [];
	}
	for (const n of body.children) {
		dfs(null, n);
	}

	return Array.from(groups.values());
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
