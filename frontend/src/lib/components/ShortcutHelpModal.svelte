<script module>
	import type { Action } from 'svelte/action';
	import { install, uninstall } from '@github/hotkey';

	let show = $state(false);
	export function toggleShow() {
		show = !show;
	}

	export const shortcuts = {
		showHelp: { keys: 'Shift+?', desc: 'Show keyboard shortcuts help' },
		nextItem: { keys: 'j', desc: 'Next item' },
		prevItem: { keys: 'k', desc: 'Previous item' },
		toggleUnread: { keys: 'm', desc: 'Unread/read item' },
		markAllasread: { keys: 'Shift+M', desc: 'Mark all items as read' },
		toggleBookmark: { keys: 'b', desc: 'Add/remove bookmark' },
		viewOriginal: { keys: 'v', desc: 'View original item' },
		nextFeed: { keys: 'Shift+J', desc: 'Next feed' },
		prevFeed: { keys: 'Shift+K', desc: 'Previous feed' },
		openSelected: { keys: 'enter', desc: 'Open selected item/feed' },
		gotoSearchPage: { keys: 'g s', desc: 'Go to search page' },
		gotoUnreadPage: { keys: 'g u', desc: 'Go to unread page' },
		gotoBookmarksPage: { keys: 'g b', desc: 'Go to bookmarks page' },
		gotoAllItemsPage: { keys: 'g a', desc: 'Go to all items page' },
		gotoFeedsPage: { keys: 'g f', desc: 'Go to feeds page' },
		gotoSettingsPage: { keys: 'g c', desc: 'Go to settings page' }
	};

	export function activateHotkey(node: HTMLElement, keys: string) {
		install(node, keys);
	}

	export function deactivateHotkey(node: HTMLElement) {
		uninstall(node);
	}

	export const hotkey: Action<HTMLElement, string> = (node, keys) => {
		$effect(() => {
			activateHotkey(node, keys);
			return () => {
				deactivateHotkey(node);
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

<dialog bind:this={modal} onclose={() => (show = false)} class="modal modal-bottom sm:modal-middle">
	<div class="modal-box">
		<form method="dialog">
			<button class="btn btn-sm btn-circle btn-ghost absolute top-2 right-2">✕</button>
		</form>

		<h3 class="text-lg font-bold">Keyboard shortcuts</h3>
		<ul class="space-y-2 py-4">
			{#each Object.entries(shortcuts) as [key, { keys, desc }]}
				<li class="hover:bg-base-200 flex items-center justify-between">
					<span class="text-sm">{desc}</span>
					<kbd class="kbd">{keys}</kbd>
				</li>
			{/each}
		</ul>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button>close</button>
	</form>
</dialog>
