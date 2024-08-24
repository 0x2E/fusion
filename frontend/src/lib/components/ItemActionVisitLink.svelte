<script lang="ts">
	import type { Item } from '$lib/api/model';
	import { ExternalLinkIcon } from 'lucide-svelte';
	import ItemActionBase from './ItemActionBase.svelte';

	export let data: Item;
	export let buttonClass = '';
	export let iconSize = 18;

	function visitLink(e: Event) {
		e.preventDefault();
		window.open(data.link, '_blank');
	}
	const icon = ExternalLinkIcon;
	const tooltip = 'Visit Original Link';
</script>

<ItemActionBase fn={visitLink} {tooltip} {buttonClass} {icon} {iconSize} />
