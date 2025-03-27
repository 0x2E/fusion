<script lang="ts">
	import type { Item } from '$lib/api/model';
	import { t } from '$lib/i18n';
	import { ExternalLink } from 'lucide-svelte';
	import { activateHotkey, deactivateHotkey, shortcuts } from './ShortcutHelpModal.svelte';

	interface Props {
		item: Item;
		enableHotkey?: boolean;
	}

	let { item, enableHotkey }: Props = $props();

	let el = $state<HTMLElement>();
	$effect(() => {
		if (!el) return;

		if (enableHotkey) {
			activateHotkey(el, shortcuts.viewOriginal.keys);
		} else {
			deactivateHotkey(el);
		}
	});
</script>

<div class="tooltip tooltip-bottom" data-tip={t('item.visit_the_original')}>
	<a href={item.link} target="_blank" bind:this={el} class="btn btn-ghost btn-square">
		<ExternalLink class="size-4" />
	</a>
</div>
