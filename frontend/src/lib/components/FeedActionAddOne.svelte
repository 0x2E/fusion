<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { checkValidity, createFeed, type FeedCreateForm } from '$lib/api/feed';
	import { allGroups } from '$lib/api/group';
	import type { Group } from '$lib/api/model';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		doneCallback: () => void;
	}

	let { doneCallback }: Props = $props();

	let step = $state(1);
	let form = $state<FeedCreateForm>({ group_id: 0, feeds: [{ name: '', link: '' }] });
	let loading = $state(false);
	let linkCandidate: { title: string; link: string }[] = $state([]);
	let groups: Group[] = $state([]);
	onMount(async () => {
		const resp = await allGroups();
		groups = resp;
	});

	// const fakeCandidates = [
	// 	{ title: 'test1', link: 'https://test1/1.xml' },
	// 	{ title: 'test2', link: 'https://test2/2.xml' }
	// ];

	async function handleAdd() {
		loading = true;
		toast.promise(checkValidity(form.feeds[0].link), {
			loading: 'Waiting for validating and sniffing ' + form.feeds[0].link,
			success: (resp) => {
				loading = false;
				if (resp.length < 1) {
					throw new Error(
						`No valid links were found for the RSS. Please check the link, or submit an RSS link directly`
					);
				}
				if (resp.length === 1) {
					if (!form.feeds[0].name) {
						form.feeds[0].name = resp[0].title;
					}
					form.feeds[0].link = resp[0].link;
					handleContinue();
				} else if (resp.length > 1) {
					linkCandidate = resp;
					step = 2;
				}
				return form.feeds[0].link + ' is valid';
			},
			error: (error) => {
				loading = false;
				return `Failed to validate ${form.feeds[0].link}: ${error}`;
			}
		});
	}

	async function handleContinue() {
		if (!form.feeds[0].name) {
			form.feeds[0].name = new URL(form.feeds[0].link).hostname;
		}
		try {
			await createFeed(form);
			toast.success('Feed has been created. Refreshing is running in the background');
			doneCallback();
		} catch (e) {
			toast.error((e as Error).message);
		}
		invalidateAll();
	}
</script>

{#if step === 1}
	<form onsubmit={handleAdd} class="flex flex-col">
		<fieldset class="fieldset">
			<legend class="fieldset-legend">Name</legend>
			<input type="text" class="input w-full" bind:value={form.feeds[0].name} />
			<p class="fieldset-label">Optional. Leave blank for automatic naming.</p>
		</fieldset>
		<fieldset class="fieldset">
			<legend class="fieldset-legend">Link</legend>
			<input type="url" class="input w-full" bind:value={form.feeds[0].link} required />
			<p class="fieldset-label">
				Either the RSS link or the website link. The server will automatically attempt to locate the
				RSS feed.
			</p>
			<p class="fieldset-label">The existing feed with the same link will be overridden.</p>
		</fieldset>
		<fieldset class="fieldset">
			<legend class="fieldset-legend">Group</legend>
			<select class="select w-full" bind:value={form.group_id} required>
				{#each groups as group}
					<option value={group.id}>{group.name}</option>
				{/each}
			</select>
		</fieldset>
		<button type="submit" disabled={loading} class="btn btn-primary mt-2 ml-auto">
			{#if loading}
				<span class="loading loading-spinner loading-sm"></span>
			{/if}
			<span> Submit </span>
		</button>
	</form>
{:else}
	<form onsubmit={handleContinue} class="flex flex-col">
		<fieldset class="fieldset">
			<legend class="fieldset-legend">Select a link</legend>
			{#each linkCandidate as l, index}
				<label class="fieldset-label">
					<input
						type="radio"
						id={String(index)}
						name="feed_link"
						value={l.link}
						onchange={() => {
							if (!form.feeds[0].name) {
								form.feeds[0].name = l.title;
							}
							form.feeds[0].link = l.link;
						}}
						class="radio radio-sm"
					/>
					{l.title} ({l.link})
				</label>
			{/each}
		</fieldset>
		<button type="submit" class="btn btn-primary mt-4 ml-auto">Confirm</button>
	</form>
{/if}
