<script module>
	import { install, uninstall } from '@github/hotkey';
	import type { Action } from 'svelte/action';

	let show = $state(false);
	export function toggleShow() {
		show = !show;
	}

	export const shortcuts = {
		showHelp: { keys: 'Shift+?', desc: t('shortcuts.show_help') },
		nextItem: { keys: 'j', desc: t('shortcuts.next_item') },
		prevItem: { keys: 'k', desc: t('shortcuts.prev_item') },
		toggleUnread: { keys: 'm', desc: t('shortcuts.toggle_unread') },
		markPageAsRead: {
			keys: 'Shift+M',
			desc: `${t('item.mark_as_read')} (${t('common.current_page')})`
		},
		toggleBookmark: { keys: 'b', desc: t('shortcuts.toggle_bookmark') },
		viewOriginal: { keys: 'v', desc: t('shortcuts.view_original') },
		nextFeed: { keys: 'Shift+J', desc: t('shortcuts.next_feed') },
		prevFeed: { keys: 'Shift+K', desc: t('shortcuts.prev_feed') },
		openSelected: { keys: 'Enter', desc: t('shortcuts.open_selected') },
		gotoSearchPage: { keys: 'g /,/', desc: t('shortcuts.goto_search_page') },
		gotoUnreadPage: { keys: 'g u', desc: t('shortcuts.goto_unread_page') },
		gotoBookmarksPage: { keys: 'g b', desc: t('shortcuts.goto_bookmarks_page') },
		gotoAllItemsPage: { keys: 'g a', desc: t('shortcuts.goto_all_items_page') },
		gotoFeedsPage: { keys: 'g f', desc: t('shortcuts.goto_feeds_page') },
		gotoSettingsPage: { keys: 'g s', desc: t('shortcuts.goto_settings_page') }
	};

	export function activateShortcut(node: HTMLElement, keys: string) {
		install(node, keys);
	}

	export function deactivateShortcut(node: HTMLElement) {
		uninstall(node);
	}

	export const shortcut: Action<HTMLElement, string> = (node, keys) => {
		$effect(() => {
			activateShortcut(node, keys);
			return () => {
				deactivateShortcut(node);
			};
		});
	};
</script>

<script lang="ts">
	import { t } from '$lib/i18n';

	let modal = $state<HTMLDialogElement>();

	$effect(() => {
		if (show) {
			modal?.showModal();
		}
	});
</script>

<dialog
	bind:this={modal}
	open={show}
	onclose={() => (show = false)}
	class="modal modal-bottom sm:modal-middle"
>
	<div class="modal-box">
		<form method="dialog">
			<button class="btn btn-sm btn-circle btn-ghost absolute top-2 right-2">✕</button>
		</form>

		<h3 class="text-lg font-bold">{t('common.shortcuts')}</h3>
		<ul class="space-y-2 py-4">
			{#each Object.entries(shortcuts) as [_, { keys, desc }]}
				<li class="hover:bg-base-200 flex items-center justify-between">
					<span class="text-sm">{desc}</span>
					<span class="space-x-1">
						{#each keys.split(',') as key, index}
							{#if index > 0}
								<span>{' or '}</span>
							{/if}
							{#each key.replaceAll('+', ' ').split(' ') as k}
								<kbd class="kbd">{k}</kbd>
							{/each}
						{/each}
					</span>
				</li>
			{/each}
		</ul>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button>close</button>
	</form>
</dialog>
