<script lang="ts">
	import type { Item } from '$lib/api/model';
	import { t } from '$lib/i18n';
	import { ExternalLink } from 'lucide-svelte';
	import { activateShortcut, deactivateShortcut, shortcuts } from './ShortcutHelpModal.svelte';

	interface Props {
		item: Item;
		enableShortcut?: boolean;
	}

	let { item, enableShortcut }: Props = $props();

	let el = $state<HTMLElement>();
	$effect(() => {
		if (!el) return;

		if (enableShortcut) {
			activateShortcut(el, shortcuts.viewOriginal.keys);
		} else {
			deactivateShortcut(el);
		}
	});
</script>

<div class="tooltip tooltip-bottom" data-tip={t('item.visit_the_original')}>
	<a href={item.link} target="_blank" bind:this={el} class="btn btn-ghost btn-square">
		<ExternalLink class="size-4" />
	</a>
</div>
