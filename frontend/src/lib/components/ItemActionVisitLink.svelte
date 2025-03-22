<script lang="ts">
	import type { Item } from '$lib/api/model';
	import { t } from '$lib/i18n';
	import { ExternalLink } from 'lucide-svelte';

	interface Props {
		data: Item;
	}

	let { data }: Props = $props();
</script>

<div class="tooltip tooltip-bottom" data-tip={t('item.visit_the_original')}>
	<a href={data.link} target="_blank" class="btn btn-ghost btn-square">
		<ExternalLink class="size-4" />
	</a>
</div>
