import DOMPurify from 'dompurify';

function sanitize(content: string, baseLink: string) {
	function joinURL(s: string | null) {
		if (!s) return '';
		try {
			// some rss's entry links are relative
			const res = new URL(s, baseLink).href;
			// console.debug(s + ' -> ' + res);
			return res;
		} catch (e) {
			console.log(e);
		}
		return s;
	}

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
				v.setAttribute(attr, joinURL(v.getAttribute(attr)));
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
	content = sanitize(content, link);
	content = embedYouTube(content, link);
	return content;
}
