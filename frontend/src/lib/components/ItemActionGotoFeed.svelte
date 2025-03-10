<script lang="ts">
	import type { Item } from '$lib/api/model';
	import { Newspaper } from 'lucide-svelte';

	interface Props {
		data: Item;
	}

	let { data }: Props = $props();
</script>

<div class="tooltip tooltip-bottom" data-tip={'Go to Feed'}>
	<a href={'/feeds/' + data.feed.id} class="btn btn-ghost btn-square">
		<Newspaper class="size-4" />
	</a>
</div>
