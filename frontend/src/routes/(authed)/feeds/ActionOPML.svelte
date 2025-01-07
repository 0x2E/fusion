<script lang="ts">
	import type { groupFeeds } from './+page';
	import * as Sheet from '$lib/components/ui/sheet';
	import * as Tabs from '$lib/components/ui/tabs';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import { createFeed } from '$lib/api/feed';
	import { toast } from 'svelte-sonner';
	import { invalidateAll } from '$app/navigation';
	import { dump, parse } from '$lib/opml';
	import { FolderIcon } from 'lucide-svelte';
	import { createGroup } from '$lib/api/group';
	import { Input } from '$lib/components/ui/input';

	interface Props {
		groups: groupFeeds[];
		open: boolean;
	}

	let { groups, open = $bindable() }: Props = $props();

	let uploadedOpmls: FileList = $state();
	let parsedGroupFeeds: { name: string; feeds: { name: string; link: string }[] }[] = $state([]);
	let importing = $state(false);

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
		let success = 0;
		const existingGroups = groups.map((v) => {
			return { id: v.id, name: v.name };
		});
		for (const g of parsedGroupFeeds) {
			try {
				let groupID = existingGroups.find((v) => v.name === g.name)?.id;
				if (groupID === undefined) {
					groupID = (await createGroup(g.name)).id;
					toast.success(`Created group ${g.name}`);
				}
				await createFeed({ group_id: groupID, feeds: g.feeds });
				toast.success(`Imported into group ${g.name}`);
				success++;
			} catch (e) {
				toast.error(`Failed to import group ${g.name}, error: ${(e as Error).message}`);
				break;
			}
		}
		if (success === parsedGroupFeeds.length) {
			toast.success('All feeds have been imported. Refreshing is running in the background');
		}
		importing = false;
		invalidateAll();
	}

	async function handleExportFeeds() {
		const data = groups.map((g) => {
			return {
				name: g.name,
				feeds: g.feeds.map((f) => {
					return { name: f.name, link: f.link };
				})
			};
		});
		const content = dump(data);
		console.log(content);
		const link = document.createElement('a');
		link.href = 'data:text/xml;charset=utf-8,' + encodeURIComponent(content);
		link.download = 'feeds.opml';
		document.body.appendChild(link);
		link.click();
		document.body.removeChild(link);
	}

	$effect(() => {
		parseOPML(uploadedOpmls);
	});

	$effect(() => {
		if (!open) {
			parsedGroupFeeds = [];
		}
	});
</script>

<Sheet.Root bind:open>
	<Sheet.Content class="w-full md:max-w-[700px] overflow-y-auto">
		<Sheet.Header>
			<Sheet.Title>Import or Export Feeds</Sheet.Title>
			<Sheet.Description>
				Feeds file should be <a
					href="http://opml.org/spec2.opml"
					target="_blank"
					class="underline font-medium">OPML</a
				> format.
			</Sheet.Description>
		</Sheet.Header>
		<Tabs.Root value="import" class="w-full mt-8">
			<Tabs.List class="mb-4">
				<Tabs.Trigger value="import">Import</Tabs.Trigger>
				<Tabs.Trigger value="export">Export</Tabs.Trigger>
			</Tabs.List>
			<Tabs.Content value="import">
				<form class="space-y-2" onsubmit={handleImportFeeds}>
					<div>
						<Label for="feed_file">File</Label>
						<input
							type="file"
							id="feed_file"
							name="feed_file"
							accept=".opml,.xml,.txt"
							required
							bind:files={uploadedOpmls}
							class="border-input placeholder:text-muted-foreground focus-visible:ring-ring flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-base shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium focus-visible:outline-none focus-visible:ring-1 disabled:cursor-not-allowed disabled:opacity-50 md:text-sm"
						/>
					</div>
					{#if parsedGroupFeeds.length > 0}
						<div>
							<p class="text-sm text-green-700">Parsed successfully.</p>
							<div
								class="p-2 rounded-md border bg-muted/40 text-muted-foreground text-nowrap overflow-x-auto"
							>
								{#each parsedGroupFeeds as group}
									<div class="flex flex-row items-center gap-1">
										<FolderIcon size={14} />{group.name}
									</div>
									<ul class="list-inside list-decimal ml-[2ch] [&:not(:last-child)]:mb-2">
										{#each group.feeds as feed}
											<li>{feed.name}, {feed.link}</li>
										{/each}
									</ul>
								{/each}
							</div>
						</div>
					{/if}
					<div class="text-sm text-secondary-foreground">
						<p>Note:</p>
						<p>
							1. Feeds will be imported into the corresponding group, which will be created
							automatically if it does not exist.
						</p>
						<p>
							2. Multidimensional group will be flattened to a one-dimensional structure, using a
							naming convention like 'a/b/c'.
						</p>
						<p>3. The existing feed with the same link will be override.</p>
					</div>
					<Button type="submit" disabled={importing}>Import</Button>
				</form>
			</Tabs.Content>
			<Tabs.Content value="export">
				<Button onclick={handleExportFeeds}>Download</Button>
			</Tabs.Content>
		</Tabs.Root>
	</Sheet.Content>
</Sheet.Root>
