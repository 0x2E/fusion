<script lang="ts">
	import { refreshFeeds } from '$lib/api/feed';
	import { toast } from 'svelte-sonner';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';

	interface Props {
		open: boolean;
	}

	let { open = $bindable() }: Props = $props();

	async function handleRefreshAll() {
		try {
			await refreshFeeds({ all: true });
			toast.success('Start refreshing in the background');
		} catch (e) {
			toast.error((e as Error).message);
		}
	}
</script>

<AlertDialog.Root bind:open closeOnOutsideClick={true}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Are you absolutely sure?</AlertDialog.Title>
			<AlertDialog.Description>
				This action will refresh <b>ALL(except suspended)</b> feeds. It may take some time.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel on:click={() => (open = false)}>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action on:click={handleRefreshAll}>Continue</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
