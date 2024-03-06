<script lang="ts">
	import type { groupFeeds } from './+page';
	import * as Sheet from '$lib/components/ui/sheet';
	import * as Select from '$lib/components/ui/select';
	import * as Tabs from '$lib/components/ui/tabs';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import { createFeed } from '$lib/api/feed';
	import { toast } from 'svelte-sonner';
	import { invalidateAll } from '$app/navigation';
	import { dump, parse } from '$lib/opml';

	export let groups: groupFeeds[];
	export let open: boolean;

	let uploadedOpmls: FileList;
	$: parseOPML(uploadedOpmls);
	let opmlGroup = { id: groups[0].id, name: groups[0].name };
	let parsedOpmlFeeds: { name: string; link: string }[] = [];
	$: {
		if (!open) {
			parsedOpmlFeeds = [];
		}
	}

	function parseOPML(opmls: FileList) {
		if (!opmls) return;
		const reader = new FileReader();
		reader.onload = (f) => {
			const content = f.target?.result?.toString();
			if (!content) {
				toast.error('Failed to load file content');
				return;
			}
			parsedOpmlFeeds = parse(content);
			console.log(parsedOpmlFeeds);
		};
		reader.readAsText(opmls[0]);
	}

	async function handleImportFeeds() {
		try {
			await createFeed({ group_id: opmlGroup.id, feeds: parsedOpmlFeeds });
			toast.success('Feeds have been imported. Refreshing is running in the background');
		} catch (e) {
			toast.error((e as Error).message);
		}
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
</script>

<Sheet.Root bind:open>
	<Sheet.Content class="w-full md:w-auto">
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
				<form class="space-y-2" on:submit|preventDefault={handleImportFeeds}>
					<div>
						<Label for="group">Group</Label>
						<Select.Root
							disabled={groups.length < 2}
							items={groups.map((v) => {
								return { value: v.id, label: v.name };
							})}
							onSelectedChange={(v) => v && (opmlGroup.id = v.value)}
						>
							<Select.Trigger>
								<Select.Value placeholder={opmlGroup.name} />
							</Select.Trigger>
							<Select.Content>
								{#each groups as g}
									<Select.Item value={g.id}>{g.name}</Select.Item>
								{/each}
							</Select.Content>
						</Select.Root>
					</div>
					<div>
						<Label for="feed_file">File</Label>
						<input
							type="file"
							id="feed_file"
							accept=".opml,.xml,.txt"
							required
							bind:files={uploadedOpmls}
							class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
						/>
					</div>
					{#if parsedOpmlFeeds.length > 0}
						<div>
							<p class="text-sm text-muted-foreground">Parsed out {parsedOpmlFeeds.length} feeds</p>
							<div
								class="max-h-[200px] overflow-scroll p-2 rounded-md border bg-muted text-muted-foreground text-nowrap"
							>
								<ul>
									{#each parsedOpmlFeeds as feed, index}
										<li>{index + 1}. <b>{feed.name}</b> {feed.link}</li>
									{/each}
								</ul>
							</div>
						</div>
					{/if}
					<Button type="submit">Import</Button>
				</form>
			</Tabs.Content>
			<Tabs.Content value="export">
				<Button on:click={handleExportFeeds}>Download</Button>
			</Tabs.Content>
		</Tabs.Root>
	</Sheet.Content>
</Sheet.Root>
