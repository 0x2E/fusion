<script lang="ts">
	import moment from 'moment';
	import DOMPurify from 'dompurify';
	import type { PageData } from './$types';
	import ItemActionFloating from '$lib/components/ItemActionFloating.svelte';
	import { Separator } from '$lib/components/ui/separator';
	import { onMount } from 'svelte';

	export let data: PageData;
	$: safeContent = sanitize(data.content, data.link);

	function sanitize(content: string, baseLink: string) {
		function joinURL(s: string | null) {
			if (!s) return '';
			try {
				// some rss's entry link is relative,
				// we cannot determine the base url
				const res = new URL(s, baseLink).href;
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

		const dom = new DOMParser().parseFromString(content, 'text/html');
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

		return DOMPurify.sanitize(replaced);
	}

	let fixActionbar = true;
	onMount(() => {
		const observer = new IntersectionObserver((entries) => {
			entries.forEach((entry) => {
				fixActionbar = !entry.isIntersecting;
			});
		});
		observer.observe(document.querySelector('#actionbar-anchor')!);
	});
</script>

<div class="max-w-prose mx-auto">
	<h1 class="text-3xl font-bold mb-4">{data.title}</h1>
	<p class="text-sm text-muted-foreground">
		<a href={'/all?feed_id=' + data.feed.id} class="hover:underline">{data.feed.name}</a> / {moment(
			data.pub_date
		).format('lll')}
	</p>

	<article
		class="mt-6 prose dark:prose-invert prose-lg text-wrap prose-pre:whitespace-pre-wrap prose-pre:break-words"
	>
		<!-- eslint-disable-next-line svelte/no-at-html-tags -->
		{@html safeContent}
	</article>

	<Separator class="my-4" />
	<p class="text-muted-foreground text-center mb-4">End of Content</p>

	<ItemActionFloating bind:data bind:fixed={fixActionbar} />
	<div id="actionbar-anchor"></div>
</div>
