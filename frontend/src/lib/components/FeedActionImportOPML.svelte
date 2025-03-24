<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { createFeed } from '$lib/api/feed';
	import { allGroups, createGroup } from '$lib/api/group';
	import type { Group } from '$lib/api/model';
	import { t } from '$lib/i18n';
	import { parse } from '$lib/opml';
	import { Folder } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		doneCallback: () => void;
	}

	let { doneCallback }: Props = $props();
	let formError = $state('');
	let importing = $state(false);
	let importLog = $state<{ content: string; isError?: boolean }[]>([]);
	let parsedGroupFeeds: { name: string; feeds: { name: string; link: string }[] }[] = $state([]);
	let uploadedOpmls = $state<FileList>();

	let groups: Group[] = $state([]);
	onMount(async () => {
		const resp = await allGroups();
		groups = resp;
	});

	$effect(() => {
		if (uploadedOpmls) {
			importLog = [];
			parseOPML(uploadedOpmls);
		}
	});

	function parseOPML(opmls: FileList) {
		formError = '';
		if (!opmls) return;

		try {
			const reader = new FileReader();
			reader.onload = (f) => {
				const content = f.target?.result?.toString();
				if (!content) {
					throw new Error(t('feed.import.opml.file_read_error'));
				}
				parsedGroupFeeds = parse(content).filter((v) => v.feeds.length > 0);
				console.log(parsedGroupFeeds);
			};
			reader.readAsText(opmls[0]);
		} catch (e) {
			formError = (e as Error).message;
		}
	}

	async function handleImportFeeds(e: Event) {
		e.preventDefault();

		importing = true;
		const existingGroups = groups.map((v) => {
			return { id: v.id, name: v.name };
		});
		for (const g of parsedGroupFeeds) {
			let groupID = existingGroups.find((v) => v.name === g.name)?.id;
			importLog.push({ content: `=== ${g.name} ===` });

			if (groupID === undefined) {
				importLog.push({ content: `➕ ${g.name}` });
				try {
					groupID = (await createGroup(g.name)).id;
				} catch (e) {
					importLog.push({
						content: (e as Error).message,
						isError: true
					});
					continue;
				}
			}
			try {
				await createFeed({ group_id: groupID, feeds: g.feeds });
				g.feeds.forEach((f) => importLog.push({ content: `✅ ${f.link}` }));
			} catch (e) {
				importLog.push({
					content: (e as Error).message,
					isError: true
				});
				g.feeds.forEach((f) => importLog.push({ content: `❌ ${f.link}`, isError: true }));
				continue;
			}
		}
		importing = false;
		if (!importLog.find((v) => v.isError)) {
			toast.success(t('state.success'));
			doneCallback();
		}
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

<form onsubmit={handleImportFeeds} class="flex flex-col">
	<fieldset class="fieldset">
		<legend class="fieldset-legend">{t('feed.import.opml.file.label')}</legend>
		<input
			type="file"
			bind:files={uploadedOpmls}
			accept=".opml,.xml,.txt"
			required
			class="file-input"
		/>
		<p class="fieldset-label">
			{@html t('feed.import.opml.file.description', {
				opml: `<a
				href="http://opml.org/spec2.opml"
				target="_blank"
				class="font-medium underline">OPML</a
			>`
			})}
		</p>
	</fieldset>
	<details>
		<summary class="text-base-content/60 text-sm font-medium">
			{t('feed.import.opml.how_it_works.title')}
		</summary>
		<div class="text-base-content/60 text-sm">
			<ul class="list-inside list-disc">
				<li>
					{t('feed.import.opml.how_it_works.description.1')}
				</li>
				<li>{t('feed.import.opml.how_it_works.description.2')}</li>
				<li>{t('feed.import.opml.how_it_works.description.3')}</li>
			</ul>
		</div>
	</details>
	{#if parsedGroupFeeds.length > 0}
		<div>
			<div class="bg-base-200 overflow-x-auto rounded-md p-2 text-sm text-nowrap">
				{#each parsedGroupFeeds as group}
					<div class="flex flex-row items-center gap-1">
						<Folder size={14} />{group.name}
					</div>
					<ul class="ml-[2ch] list-inside list-decimal [&:not(:last-child)]:mb-2">
						{#each group.feeds as feed}
							<li>
								{feed.name} (<a href={feed.link} target="_blank" class="text-blue-600"
									>{feed.link}</a
								>)
							</li>
						{/each}
					</ul>
				{/each}
			</div>
			<ul class="mt-2 list-inside list-disc">
				{#each importLog as log}
					<li class={log.isError ? 'text-error' : ''}>{log.content}</li>
				{/each}
			</ul>
		</div>
	{/if}

	<button type="submit" disabled={importing} class="btn btn-primary mt-4 ml-auto">
		{#if importing}
			<span class="loading loading-spinner loading-sm"></span>
		{/if}
		<span>{t('common.submit')}</span>
	</button>
</form>
