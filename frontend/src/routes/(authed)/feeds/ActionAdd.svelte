<script lang="ts">
	import { Loader2Icon } from 'lucide-svelte';
	import type { Feed } from '$lib/api/model';
	import type { groupFeeds } from './+page';
	import * as RadioGroup from '$lib/components/ui/radio-group';
	import * as Sheet from '$lib/components/ui/sheet';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Select from '$lib/components/ui/select';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { checkValidity, createFeed } from '$lib/api/feed';
	import { toast } from 'svelte-sonner';
	import { invalidateAll } from '$app/navigation';

	export let groups: groupFeeds[];
	export let open: boolean;

	function emptyForm(): Feed {
		return {
			id: 0,
			name: '',
			link: '',
			failure: '',
			updated_at: new Date(),
			suspended: false,
			group: { id: groups[0].id, name: groups[0].name }
		};
	}

	let openCandidate = false;
	let loading = false;
	let formData = emptyForm();
	let linkCandidate: { title: string; link: string }[] = [];

	$: {
		if (!open) {
			formData = emptyForm();
		}
	}

	async function handleAdd() {
		loading = true;
		toast.promise(checkValidity(formData.link), {
			loading: 'Waiting for validating and sniffing ' + formData.link,
			success: (resp) => {
				loading = false;
				// resp = [
				// 	{ title: 'test1', link: 'https://test1/1.xml' },
				// 	{ title: 'test2', link: 'https://test2/2.xml' }
				// ];
				if (resp.length < 1) {
					throw new Error(
						`No valid links were found for the RSS. Please check the link, or submit an RSS link directly`
					);
				}
				if (resp.length === 1) {
					formData.link = resp[0].link;
					handleContinue();
				} else if (resp.length > 1) {
					linkCandidate = resp;
					openCandidate = true;
				}
				return formData.link + ' is valid';
			},
			error: (error) => {
				loading = false;
				return `Failed to validate ${formData.link}: ${error}`;
			}
		});
	}

	async function handleContinue() {
		openCandidate = false;
		try {
			await createFeed({ group_id: formData.group.id, feeds: [formData] });
			toast.success(formData.name + ' has been created. Refreshing is running in the background');
		} catch (e) {
			toast.error((e as Error).message);
		}
		invalidateAll();
		open = false;
	}
</script>

<Sheet.Root bind:open>
	<Sheet.Content class="w-full md:min-w-[500px] md:w-auto">
		<Sheet.Header>
			<Sheet.Title>Add feed</Sheet.Title>
			<Sheet.Description>
				Enter the direct RSS link. Or a website link to find its RSS links automatically.
			</Sheet.Description>
		</Sheet.Header>
		<div class="w-full mt-4">
			<form on:submit|preventDefault={handleAdd} class="flex flex-col gap-2">
				<div>
					<Label for="name">Name</Label>
					<Input
						id="name"
						type="text"
						class="w-full"
						on:change={(e) => {
							if (e.target instanceof HTMLInputElement) {
								formData.name = e.target.value;
							}
						}}
						required
					/>
					<p class="text-sm text-muted-foreground">
						The existing feed with the same link will be override.
					</p>
				</div>

				<div>
					<Label for="link">Link</Label>
					<Input
						id="link"
						type="text"
						class="w-full"
						on:change={(e) => {
							if (e.target instanceof HTMLInputElement) {
								formData.link = e.target.value;
							}
						}}
						required
					/>
				</div>

				<div>
					<Label for="group" class="mt-4">Group</Label>
					<Select.Root
						disabled={groups.length < 2}
						items={groups.map((v) => {
							return { value: v.id, label: v.name };
						})}
						onSelectedChange={(v) => v && (formData.group.id = v.value)}
					>
						<Select.Trigger>
							<Select.Value placeholder={formData.group.name} />
						</Select.Trigger>
						<Select.Content>
							{#each groups as g}
								<Select.Item value={g.id}>{g.name}</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>

				<Button type="submit" class="mt-8" disabled={loading}>
					{#if loading}
						<Loader2Icon class="mr-2 h-4 w-4 animate-spin" />
					{:else}
						Save
					{/if}
				</Button>
			</form>
		</div>
	</Sheet.Content>
</Sheet.Root>

<Dialog.Root bind:open={openCandidate}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Links Selection</Dialog.Title>
			<Dialog.Description>
				More than one links have been found automatically. Please choose one.
			</Dialog.Description>
		</Dialog.Header>
		<RadioGroup.Root onValueChange={(v) => (formData.link = v)}>
			{#each linkCandidate as link}
				<div class="flex items-center space-x-2">
					<RadioGroup.Item value={link.link} id={link.link} />
					<Label for={link.link}><b>{link.title}</b>: {link.link}</Label>
				</div>
			{/each}
		</RadioGroup.Root>
		<Dialog.Footer>
			<Button variant="secondary" on:click={() => (openCandidate = false)}>Cancel</Button>
			<Button on:click={() => handleContinue()}>Continue</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
