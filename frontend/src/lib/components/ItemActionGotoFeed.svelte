<script lang="ts">
	import type { Item } from '$lib/api/model';
	import { t } from '$lib/i18n';
	import { Newspaper } from 'lucide-svelte';

	interface Props {
		data: Item;
	}

	let { data }: Props = $props();
</script>

<div class="tooltip tooltip-bottom" data-tip={t('item.goto_feed')}>
	<a href={'/feeds/' + data.feed.id} class="btn btn-ghost btn-square">
		<Newspaper class="size-4" />
	</a>
</div>
