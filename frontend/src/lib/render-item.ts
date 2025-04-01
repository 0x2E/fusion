import DOMPurify from 'dompurify';
import { tryAbsURL } from './utils';

function sanitize(content: string, baseLink: string) {
	const elements: { tag: string; attrs: string[] }[] = [
		{ tag: 'a', attrs: ['href'] },
		{ tag: 'img', attrs: ['src'] }, //TODO: srcset attr and base64 type img
		{ tag: 'audio', attrs: ['src'] },
		{ tag: 'source', attrs: ['src'] },
		{ tag: 'video', attrs: ['src'] },
		{ tag: 'embed', attrs: ['src'] },
		{ tag: 'object', attrs: ['data'] }
	];

	const cleaned = DOMPurify.sanitize(content, { FORBID_ATTR: ['class', 'style'] });

	const dom = new DOMParser().parseFromString(cleaned, 'text/html');
	for (const el of elements) {
		dom.querySelectorAll(el.tag).forEach((v) => {
			for (const attr of el.attrs) {
				const link = v.getAttribute(attr);
				if (!link) continue;
				v.setAttribute(attr, tryAbsURL(link, baseLink));
			}
		});
	}

	// prevent table from overflowing
	// https://github.com/tailwindlabs/tailwindcss-typography/issues/334#issuecomment-1942177668
	dom.querySelectorAll('table').forEach((v) => {
		if (v.parentNode) {
			const parentDiv = document.createElement('div');
			parentDiv.classList.add('overflow-x-auto');
			v.parentNode.insertBefore(parentDiv, v);
			parentDiv.appendChild(v);
		}
	});

	// data.content = data.content.replace(/src="(.*?)"/g, (_, match) => {
	// 	const res = new URL(match, data.link).href;
	// 	return `src="${res}"`;
	// });

	return new XMLSerializer().serializeToString(dom);
}

function embedYouTube(content: string, link: string): string {
	const youtubeDomains = ['youtube.com', 'youtu.be'];
	if (youtubeDomains.find((v) => new URL(link).hostname.endsWith(v))) {
		const videoID = new URL(link).searchParams.get('v');
		content =
			`<iframe style="aspect-ratio: 16 / 9; width: 100% !important;" src="http://www.youtube.com/embed/` +
			videoID +
			`" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>` +
			content;
	}
	return content;
}

export function render(content: string, link: string): string {
	link = tryAbsURL(link);
	content = sanitize(content, link);
	content = embedYouTube(content, link);
	return content;
}
