<script lang="ts">
	import ItemActionBookmark from '$lib/components/ItemActionBookmark.svelte';
	import ItemActionGotoFeed from '$lib/components/ItemActionGotoFeed.svelte';
	import ItemActionUnread from '$lib/components/ItemActionUnread.svelte';
	import ItemActionVisitLink from '$lib/components/ItemActionVisitLink.svelte';
	import PageNavHeader from '$lib/components/PageNavHeader.svelte';
	import DOMPurify from 'dompurify';
	import { ExternalLink } from 'lucide-svelte';
	import ItemSwitcher from './ItemSwitcher.svelte';

	let { data } = $props();

	function sanitize(content: string, baseLink: string) {
		function joinURL(s: string | null) {
			if (!s) return '';
			try {
				// some rss's entry link is relative,
				// we cannot determine the base url
				const res = new URL(s, baseLink).href;
				console.debug(s + ' -> ' + res);
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

	let safeContent = $derived(sanitize(data.content, data.link));
</script>

<PageNavHeader title={data.title}>
	<ItemActionGotoFeed {data} />
	<ItemActionUnread {data} />
	<ItemActionBookmark {data} />
	<ItemActionVisitLink {data} />
</PageNavHeader>

<div class="relative flex h-full w-full justify-around px-4 py-6">
	<ItemSwitcher itemID={data.id} action="previous" />
	<article class="w-full max-w-prose">
		<p class="text-base-content/60 flex flex-col text-sm md:flex-row">
			{new Date(data.pub_date).toLocaleString()}
		</p>

		<div class="prose text-wrap break-words">
			<h1>
				<a
					href={data.link}
					target="_blank"
					class="inline-flex items-center gap-2 no-underline hover:underline"
				>
					<span>
						{data.title}
					</span>
					<ExternalLink class="hidden size-5 md:block" />
				</a>
			</h1>
			{@html safeContent}
		</div>
	</article>
	<ItemSwitcher itemID={data.id} action="next" />
</div>
