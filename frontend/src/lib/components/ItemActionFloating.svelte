<script lang="ts">
	import type { Item } from '$lib/api/model';
	import { ArrowUp, ArrowUpIcon } from 'lucide-svelte';
	import ItemActionBookmark from './ItemActionBookmark.svelte';
	import ItemActionNavigate from './ItemActionNavigate.svelte';
	import ItemActionUnread from './ItemActionUnread.svelte';
	import ItemActionVisitLink from './ItemActionVisitLink.svelte';
	import { Separator } from './ui/separator';

	interface Props {
		data: Item;
		fixed?: boolean;
	}

	let { data, fixed = true }: Props = $props();

	function handleScrollTop(e: Event) {
		e.preventDefault();
		document.body.scrollIntoView({ behavior: 'smooth' });
	}
</script>

<div class="{fixed ? 'fixed' : ''} bg-base-100 right-0 bottom-2 left-0">
	<div
		class="bg-background mx-auto flex w-fit flex-row items-center justify-center gap-2 rounded-full border px-6 py-2 shadow"
	>
		<ItemActionUnread {data} />
		<Separator orientation="vertical" class="h-5" />
		<ItemActionBookmark {data} />
		<Separator orientation="vertical" class="h-5" />
		<ItemActionVisitLink {data} />
		<Separator orientation="vertical" class="h-5" />
		<div class="tooltip" data-tip={'Back to Top'}>
			<button onclick={handleScrollTop} class="btn btn-ghost btn-square btn-sm">
				<ArrowUp class="size-5" />
			</button>
		</div>

		<Separator orientation="vertical" class="h-5" />
		<Separator orientation="vertical" class="h-5" />
		<ItemActionNavigate {data} />
	</div>
</div>
