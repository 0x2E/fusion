<script lang="ts">
	import type { Item } from '$lib/api/model';
	import { ExternalLink } from 'lucide-svelte';

	interface Props {
		data: Item;
	}

	let { data }: Props = $props();
</script>

<div class="tooltip tooltip-bottom" data-tip={'Visit Original Link'}>
	<a href={data.link} target="_blank" class="btn btn-ghost btn-square">
		<ExternalLink class="size-4" />
	</a>
</div>
