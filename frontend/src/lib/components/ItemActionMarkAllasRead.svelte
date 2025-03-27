<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { page } from '$app/state';
	import { updateUnread } from '$lib/api/item';
	import type { Item } from '$lib/api/model';
	import { t } from '$lib/i18n';
	import { CheckCheck } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { hotkey, shortcuts } from './ShortcutHelpModal.svelte';

	type Props =
		| {
				disabled: true;
		  }
		| {
				disabled?: false;
				items: Item[];
		  };

	let props: Props = $props();

	async function handleMarkAllAsRead() {
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
</script>

<div
	class="tooltip tooltip-bottom"
	data-tip={props.disabled ? undefined : t('item.mark_all_as_read')}
>
	<button
		disabled={props.disabled}
		onclick={handleMarkAllAsRead}
		use:hotkey={shortcuts.markAllasread.keys}
		class="btn btn-ghost btn-square"
	>
		<CheckCheck class="size-4" />
	</button>
</div>
