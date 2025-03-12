<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { createFeed } from '$lib/api/feed';
	import { allGroups, createGroup } from '$lib/api/group';
	import type { Group } from '$lib/api/model';
	import { parse } from '$lib/opml';
	import { Folder } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		doneCallback: () => void;
	}

	let { doneCallback }: Props = $props();
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
		if (!opmls) return;

		const reader = new FileReader();
		reader.onload = (f) => {
			const content = f.target?.result?.toString();
			if (!content) {
				toast.error('Failed to load file content');
				return;
			}
			parsedGroupFeeds = parse(content).filter((v) => v.feeds.length > 0);
			console.log(parsedGroupFeeds);
		};
		reader.readAsText(opmls[0]);
	}

	async function handleImportFeeds(e: Event) {
		e.preventDefault();

		importing = true;
		const existingGroups = groups.map((v) => {
			return { id: v.id, name: v.name };
		});
		for (const g of parsedGroupFeeds) {
			let groupID = existingGroups.find((v) => v.name === g.name)?.id;

			if (groupID === undefined) {
				try {
					groupID = (await createGroup(g.name)).id;
					importLog.push({ content: `Created group ${g.name}` });
				} catch (e) {
					importLog.push({
						content: `Failed to create group ${g.name}. error: ${(e as Error).message}`,
						isError: true
					});
					continue;
				}
			}
			try {
				await createFeed({ group_id: groupID, feeds: g.feeds });
				g.feeds.forEach((f) => importLog.push({ content: `Imported ${f.link}` }));
			} catch (e) {
				g.feeds.forEach((f) =>
					importLog.push({
						content: `Failed to import ${g.name}. error: ${(e as Error).message}`,
						isError: true
					})
				);
				continue;
			}
		}
		importing = false;
		if (!importLog.find((v) => v.isError)) {
			toast.success('Imported successfully');
			doneCallback();
		}
		invalidateAll();
	}
</script>

<form onsubmit={handleImportFeeds} class="flex flex-col">
	<fieldset class="fieldset">
		<legend class="fieldset-legend">Pick a OPML file</legend>
		<input
			type="file"
			bind:files={uploadedOpmls}
			accept=".opml,.xml,.txt"
			required
			class="file-input"
		/>
		<p class="fieldset-label">
			The file should be <a
				href="http://opml.org/spec2.opml"
				target="_blank"
				class="font-medium underline">OPML</a
			> format. You can get one from your previous RSS reader.
		</p>
	</fieldset>
	<details>
		<summary class="text-base-content/60 text-sm font-medium"> How it works? </summary>
		<div class="text-base-content/60 text-sm">
			<ul class="list-inside list-disc">
				<li>
					Feeds will be imported into the corresponding group, which will be created automatically
					if it does not exist.
				</li>
				<li>
					Multidimensional group will be flattened to a one-dimensional structure, using a naming
					convention like 'a/b/c'.
				</li>
				<li>The existing feed with the same link will be overridden.</li>
			</ul>
		</div>
	</details>
	{#if parsedGroupFeeds.length > 0}
		<div>
			<p class="text-success text-sm">Parsed successfully.</p>
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
		<span>Submit</span>
	</button>
</form>
