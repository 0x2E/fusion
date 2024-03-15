<script lang="ts">
	import moment from 'moment';
	import DOMPurify from 'dompurify';
	import type { PageData } from './$types';
	import ItemAction from '$lib/components/ItemAction.svelte';

	export let data: PageData;

	function joinURL(s: string | null) {
		if (!s) return '';
		try {
			// some rss's entry link is relative,
			// we cannot determine the base url
			const res = new URL(s, data.link).href;
			console.log(s + ' -> ' + res);
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

	const dom = new DOMParser().parseFromString(data.content, 'text/html');
	for (const el of elements) {
		dom.querySelectorAll(el.tag).forEach((v) => {
			for (const attr of el.attrs) {
				v.setAttribute(attr, joinURL(v.getAttribute(attr)));
			}
		});
	}
	const replaced = new XMLSerializer().serializeToString(dom);
	// data.content = data.content.replace(/src="(.*?)"/g, (_, match) => {
	// 	const res = new URL(match, data.link).href;
	// 	return `src="${res}"`;
	// });

	data.content = DOMPurify.sanitize(replaced);
</script>

<svelte:head>
	<title>{data.title}</title>
</svelte:head>

<div class="max-w-prose mx-auto">
	<h1 class="text-3xl font-bold mb-4">{data.title}</h1>
	<p class="text-sm text-muted-foreground">
		<a href={'/all?feed_id=' + data.feed.id} class="hover:underline">{data.feed.name}</a> / {moment(
			data.pub_date
		).format('lll')}
	</p>
	<ItemAction bind:data />

	<!-- FIX: pre overflow: https://github.com/tailwindlabs/tailwindcss-typography/issues/96 -->
	<article class="mt-6 mx-auto prose dark:prose-invert prose-lg">
		<!-- eslint-disable-next-line svelte/no-at-html-tags -->
		{@html data.content}
	</article>
</div>
