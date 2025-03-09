<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { deleteFeed, refreshFeeds, updateFeed, type FeedUpdateForm } from '$lib/api/feed';
	import type { Feed } from '$lib/api/model';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import { Separator } from '$lib/components/ui/separator';
	import * as Sheet from '$lib/components/ui/sheet';
	import { AlertCircleIcon, Loader2Icon } from 'lucide-svelte';
	import moment from 'moment';
	import { toast } from 'svelte-sonner';
	import type { groupFeeds } from './+page';

	interface Props {
		groups: groupFeeds[];
		show?: boolean;
		selectedFeed: Feed;
	}

	let { groups, show = $bindable(false), selectedFeed }: Props = $props();
	let formData: FeedUpdateForm = $state();
	let refreshing = $state(false);

	$effect(() => {
		if (show) {
			formData = {};
		}
	});

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

	async function handleUpdate(e: Event) {
		e.preventDefault();

		console.log(formData);
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
</script>

<Sheet.Root bind:open={show}>
	<Sheet.Content class="w-full overflow-scroll md:w-auto md:min-w-[500px]">
		<Sheet.Header>
			<Sheet.Title>{selectedFeed.name}</Sheet.Title>
			<Sheet.Description>
				<p>Last refreshed at {moment(selectedFeed.updated_at).format('lll')}</p>
				{#if selectedFeed.failure}
					<Alert.Root variant="destructive" class="container break-all">
						<AlertCircleIcon class="h-4 w-4" />
						<Alert.Title>Error</Alert.Title>
						<Alert.Description>{selectedFeed.failure}</Alert.Description>
					</Alert.Root>
				{/if}
			</Sheet.Description>
		</Sheet.Header>
		<div class="mt-4 flex w-full flex-col">
			{#if selectedFeed !== undefined}
				<form onsubmit={handleUpdate} class="flex flex-col gap-2">
					<div>
						<Label for="name">Name</Label>
						<Input
							id="name"
							type="text"
							class="w-full"
							value={selectedFeed.name}
							onchange={(e) => {
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
							onchange={(e) => {
								if (e.target instanceof HTMLInputElement) {
									formData.link = e.target.value;
								}
							}}
							required
						/>
					</div>
					<div>
						<Label for="req_proxy" class="mt-2">Proxy</Label>
						<Input
							id="req_proxy"
							type="text"
							class="w-full"
							value={selectedFeed.req_proxy}
							onchange={(e) => {
								if (e.target instanceof HTMLInputElement) {
									formData.req_proxy = e.target.value;
								}
							}}
						/>
						<p class="text-muted-foreground text-sm">
							Proxy for HTTP client. The types 'http', 'https', and 'socks5' are supported.
						</p>
					</div>
					<div>
						<Label for="group">Group</Label>
						<Select.Root
							type="single"
							disabled={groups.length < 2}
							onValueChange={(v) => (formData.group_id = parseInt(v))}
						>
							<Select.Trigger>
								{selectedFeed.group.name}
							</Select.Trigger>
							<Select.Content>
								{#each groups as g}
									<Select.Item value={String(g.id)}>{g.name}</Select.Item>
								{/each}
							</Select.Content>
						</Select.Root>
					</div>
					<Button class="mt-4" type="submit">Save</Button>
				</form>
			{/if}

			<Separator class="my-6" />
			<div class="flex w-full flex-col gap-2">
				<Button variant="outline" href={'/all?feed_id=' + selectedFeed.id}>View items</Button>
				<Button
					variant="secondary"
					onclick={handleRefresh}
					disabled={refreshing || selectedFeed.suspended}
				>
					{#if refreshing}
						<Loader2Icon class="mr-2 h-4 w-4 animate-spin" />
					{:else}
						Refresh
					{/if}
				</Button>
			</div>

			<Separator class="my-6" />
			<div class="flex w-full flex-col gap-2">
				<Button
					variant="secondary"
					onclick={() => {
						const alertText = `Are you sure to ${selectedFeed.suspended ? 'resume' : 'suspend'} [${selectedFeed.name}]?`;
						if (!confirm(alertText)) {
							return;
						}
						handleToggleSuspended();
					}}
				>
					{#if selectedFeed.suspended}
						Resume
					{:else}
						Suspend
					{/if}
				</Button>

				<Button
					variant="destructive"
					onclick={() => {
						if (!confirm(`Are you sure to permanently delete [${selectedFeed.name}]?`)) {
							return;
						}
						handleDelete();
					}}
				>
					Delete
				</Button>
			</div>
		</div>
	</Sheet.Content>
</Sheet.Root>
