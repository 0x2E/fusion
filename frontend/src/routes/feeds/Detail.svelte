<script lang="ts">
	import { Separator } from '$lib/components/ui/separator';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import * as Select from '$lib/components/ui/select';
	import * as Sheet from '$lib/components/ui/sheet';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { AlertCircleIcon, Loader2Icon } from 'lucide-svelte';
	import * as Alert from '$lib/components/ui/alert';
	import { deleteFeed, refreshFeeds, updateFeed, type FeedUpdateForm } from '$lib/api/feed';
	import type { Feed } from '$lib/api/model';
	import type { groupFeeds } from './+page';
	import { invalidateAll } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { Button } from '$lib/components/ui/button';
	import moment from 'moment';

	export let groups: groupFeeds[];
	export let show = false;
	export let selectedFeed: Feed;
	let formData: FeedUpdateForm;
	let refreshing = false;

	$: {
		if (show) {
			formData = {};
		}
	}

	async function handleRefresh() {
		refreshing = true;
		toast.promise(refreshFeeds({ id: selectedFeed.id }), {
			loading: 'Refreshing ' + selectedFeed.name,
			success: () => {
				invalidateAll();
				refreshing = false;
				return selectedFeed.name + ' has been refreshed';
			},
			error: (error) => {
				invalidateAll();
				refreshing = false;
				return 'Failed to refresh: ' + error;
			}
		});
	}

	async function handleDelete() {
		try {
			await deleteFeed(selectedFeed.id);
			toast.success('Feed has been deleted');
		} catch (e) {
			toast.error((e as Error).message);
		}
		show = false;
		invalidateAll();
	}

	async function handleUpdate() {
		toast.promise(updateFeed(selectedFeed.id, formData), {
			loading: 'Updating',
			success: () => {
				invalidateAll();
				return 'Update successfully';
			},
			error: (e) => {
				invalidateAll();
				return (e as Error).message;
			}
		});
	}

	async function handleToggleSuspended() {
		const data: FeedUpdateForm = { suspended: !selectedFeed.suspended };
		try {
			await updateFeed(selectedFeed.id, data);
			toast.success('Update successfully');
			invalidateAll();
		} catch (e) {
			invalidateAll();
			toast.error((e as Error).message);
		}
	}
</script>

<Sheet.Root bind:open={show}>
	<Sheet.Content class="w-full md:min-w-[500px] md:w-auto">
		<Sheet.Header>
			<Sheet.Title>{selectedFeed.name}</Sheet.Title>
			<Sheet.Description>
				<p>Last refreshed at {moment(selectedFeed.updated_at).format('lll')}</p>
				{#if selectedFeed.failure}
					<Alert.Root variant="destructive" class="container">
						<AlertCircleIcon class="h-4 w-4" />
						<Alert.Title>Error</Alert.Title>
						<Alert.Description>{selectedFeed.failure}</Alert.Description>
					</Alert.Root>
				{/if}
			</Sheet.Description>
		</Sheet.Header>
		<div class="flex flex-col w-full mt-4">
			{#if selectedFeed !== undefined}
				<form on:submit|preventDefault={handleUpdate} class="flex flex-col gap-1">
					<div>
						<Label for="name">Name</Label>
						<Input
							id="name"
							type="text"
							class="w-full"
							value={selectedFeed.name}
							on:input={(e) => {
								// two-way bind not works, so do this. https://stackoverflow.com/questions/60825553/svelte-input-binding-breaks-when-a-reactive-value-is-a-reference-type
								if (e.target instanceof HTMLInputElement) {
									formData.name = e.target.value;
								}
							}}
							required
						/>
					</div>
					<div>
						<Label for="link" class="mt-2">Link</Label>
						<Input
							id="link"
							type="text"
							class="w-full"
							value={selectedFeed.link}
							on:input={(e) => {
								if (e.target instanceof HTMLInputElement) {
									formData.link = e.target.value;
								}
							}}
							required
						/>
					</div>
					<div>
						<Label for="group">Group</Label>
						<Select.Root
							disabled={groups.length < 2}
							items={groups.map((v) => {
								return { value: v.id, label: v.name };
							})}
							onSelectedChange={(v) => v && (formData.group_id = v.value)}
						>
							<Select.Trigger>
								<Select.Value placeholder={selectedFeed.group.name} />
							</Select.Trigger>
							<Select.Content>
								{#each groups as g}
									<Select.Item value={g.id}>{g.name}</Select.Item>
								{/each}
							</Select.Content>
						</Select.Root>
					</div>
					<Button class="mt-4" type="submit">Save</Button>
				</form>
			{/if}

			<Separator class="my-6" />
			<Button
				variant="secondary"
				on:click={handleRefresh}
				disabled={refreshing || selectedFeed.suspended}
			>
				{#if refreshing}
					<Loader2Icon class="mr-2 h-4 w-4 animate-spin" />
				{:else}
					Refresh
				{/if}
			</Button>

			<Separator class="my-6" />
			<div class="flex flex-col w-full gap-2">
				<AlertDialog.Root>
					<AlertDialog.Trigger asChild let:builder>
						<Button builders={[builder]} variant="secondary">
							{#if selectedFeed.suspended}
								Resume
							{:else}
								Suspend
							{/if}
						</Button>
					</AlertDialog.Trigger>
					<AlertDialog.Content>
						<AlertDialog.Header>
							<AlertDialog.Title>Are you absolutely sure?</AlertDialog.Title>
							<AlertDialog.Description>
								{#if selectedFeed.suspended}
									This will resume the refreshing of <b>{selectedFeed.name}</b>.
								{:else}
									This will suspend the refreshing of <b>{selectedFeed.name}</b> untill you resume it.
								{/if}
							</AlertDialog.Description>
						</AlertDialog.Header>
						<AlertDialog.Footer>
							<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
							<AlertDialog.Action on:click={handleToggleSuspended}>Continue</AlertDialog.Action>
						</AlertDialog.Footer>
					</AlertDialog.Content>
				</AlertDialog.Root>

				<AlertDialog.Root>
					<AlertDialog.Trigger asChild let:builder>
						<Button builders={[builder]} variant="destructive">Delete</Button>
					</AlertDialog.Trigger>
					<AlertDialog.Content>
						<AlertDialog.Header>
							<AlertDialog.Title>Are you absolutely sure?</AlertDialog.Title>
							<AlertDialog.Description>
								This will permanently delete <b>{selectedFeed.name}</b>
								and its items.
							</AlertDialog.Description>
						</AlertDialog.Header>
						<AlertDialog.Footer>
							<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
							<AlertDialog.Action on:click={handleDelete}>Continue</AlertDialog.Action>
						</AlertDialog.Footer>
					</AlertDialog.Content>
				</AlertDialog.Root>
			</div>
		</div>
	</Sheet.Content>
</Sheet.Root>
