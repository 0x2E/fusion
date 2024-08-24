<script lang="ts">
	import { ArrowUpIcon } from 'lucide-svelte';
	import type { Item } from '$lib/api/model';
	import ItemActionUnread from './ItemActionUnread.svelte';
	import ItemActionBookmark from './ItemActionBookmark.svelte';
	import ItemActionVisitLink from './ItemActionVisitLink.svelte';
	import ItemActionBase from './ItemActionBase.svelte';

	export let data: Item;
	export let fixed = true;

	function handleScrollTop(e: Event) {
		e.preventDefault();
		document.body.scrollIntoView({ behavior: 'smooth' });
	}
</script>

<div class="{fixed ? 'fixed' : ''} bottom-2 left-0 right-0">
	<div
		class="flex flex-row justify-center items-center gap-2 rounded-full border shadow w-fit mx-auto bg-background px-6 py-2"
	>
		<ItemActionUnread {data} />
		<ItemActionBookmark {data} />
		<ItemActionVisitLink {data} />
		<ItemActionBase fn={handleScrollTop} tooltip="Back to Top" icon={ArrowUpIcon} />
	</div>
</div>
