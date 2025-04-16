<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { page } from '$app/state';
	import { listItems, parseURLtoFilter, updateUnread } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { t } from '$lib/i18n';
	import { CheckCheck } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { shortcut, shortcuts } from './ShortcutHelpModal.svelte';

	type Props =
		| {
				disabled: true;
		  }
		| {
				disabled?: false;
				items: Item[];
		  };

	let props: Props = $props();

	async function handleMarkPageAsRead() {
		if (props.disabled) {
			console.error('unreachable code');
			return;
		}

		try {
			const ids = props.items.map((v) => v.id);
			await updateUnread(ids, false);
			toast.success(t('state.success'));
			invalidate('page:' + page.url.pathname);
		} catch (e) {
			toast.error((e as Error).message);
		}
	}

	async function handleMarkAllAsRead() {
		if (props.disabled) {
			console.error('unreachable code');
			return;
		}

		const feed_id = props.items.at(0)?.feed.id;
		if (!feed_id) {
			console.error('unreachable code');
			return;
		}

		try {
			while (true) {
				const resp = await listItems({ feed_id: feed_id, page: 1, page_size: 200, unread: true });
				if (resp.items.length === 0) {
					break;
				}
				const ids = resp.items.map((v) => v.id);
				await updateUnread(ids, false);
			}
			toast.success(t('state.success'));
			invalidate('page:' + page.url.pathname);
		} catch (e) {
			toast.error((e as Error).message);
		}
	}
</script>

<div class="tooltip tooltip-bottom" data-tip={props.disabled ? undefined : t('item.mark_as_read')}>
	<details class="dropdown dropdown-end">
		<summary class="btn btn-ghost btn-square">
			<CheckCheck class="size-4" />
		</summary>
		<ul class="menu dropdown-content bg-base-100 rounded-box z-1 w-52 p-2 shadow-sm">
			<li class="menu-title text-xs">{t('item.mark_as_read')}</li>
			<li>
				<button
					disabled={props.disabled}
					onclick={handleMarkPageAsRead}
					use:shortcut={shortcuts.markPageAsRead.keys}
				>
					{t('common.current_page')}
				</button>
			</li>
			<li>
				<button disabled={props.disabled} onclick={handleMarkAllAsRead}>
					{t('common.all')}
				</button>
			</li>
		</ul>
	</details>
</div>
