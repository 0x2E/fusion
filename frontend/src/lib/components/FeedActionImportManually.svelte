<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { checkValidity, createFeed, type FeedCreateForm } from '$lib/api/feed';
	import { allGroups } from '$lib/api/group';
	import type { Group } from '$lib/api/model';
	import { t } from '$lib/i18n';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		doneCallback: () => void;
	}

	let { doneCallback }: Props = $props();

	let step = $state(1);
	let form = $state<FeedCreateForm>({ group_id: 1, feeds: [{ name: '', link: '' }] });
	let formError = $state('');
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
		formError = '';
		loading = true;
		try {
			const resp = await checkValidity(form.feeds[0].link);
			loading = false;
			if (resp.length < 1) {
				throw new Error(t('feed.import.manually.no_valid_feed_error'));
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
			return;
		} catch (e) {
			loading = false;
			formError = (e as Error).message;
		}
	}

	async function handleContinue() {
		if (!form.feeds[0].name) {
			form.feeds[0].name = new URL(form.feeds[0].link).hostname;
		}
		try {
			await createFeed(form);
			toast.success(t('state.success'));
			doneCallback();
		} catch (e) {
			formError = (e as Error).message;
		}
		loading = false;
		invalidateAll();
	}
</script>

{#if formError}
	<div role="alert" class="alert alert-error">
		<svg
			xmlns="http://www.w3.org/2000/svg"
			class="h-6 w-6 shrink-0 stroke-current"
			fill="none"
			viewBox="0 0 24 24"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
			/>
		</svg>
		<span>{formError}</span>
	</div>
{/if}

{#if step === 1}
	<form onsubmit={handleAdd} class="flex flex-col">
		<fieldset class="fieldset">
			<legend class="fieldset-legend">{t('common.link')}</legend>
			<input type="url" class="input w-full" bind:value={form.feeds[0].link} required />
			<p class="fieldset-label">
				{t('feed.import.manually.link.description')}
			</p>
		</fieldset>
		<fieldset class="fieldset">
			<legend class="fieldset-legend">{t('common.name')}</legend>
			<input type="text" class="input w-full" bind:value={form.feeds[0].name} />
			<p class="fieldset-label">{t('feed.import.manually.name.description')}</p>
		</fieldset>
		<fieldset class="fieldset">
			<legend class="fieldset-legend">{t('common.group')}</legend>
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
			<span> {t('common.submit')} </span>
		</button>
	</form>
{:else}
	<form onsubmit={handleContinue} class="flex flex-col">
		<fieldset class="fieldset">
			<legend class="fieldset-legend">{t('feed.import.manually.link_candidates.label')}</legend>
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
		<button type="submit" class="btn btn-primary mt-4 ml-auto">{t('common.confirm')}</button>
	</form>
{/if}
