<script lang="ts">
	import type { PageData } from './$types';
	import type { Feed } from '$lib/api/model';
	import { Button } from '$lib/components/ui/button';
	import * as Tabs from '$lib/components/ui/tabs';
	import PageHead from '$lib/components/PageHead.svelte';
	import Detail from './Detail.svelte';
	import Actions from './Actions.svelte';
	import { AlertCircleIcon, CirclePauseIcon } from 'lucide-svelte';

	export let data: PageData;

	let showDetail = false;
	let currentGroup = 1;
	let selectedFeed: Feed;

	function handleShowDetail(f: Feed) {
		showDetail = true;
		selectedFeed = f;
	}
</script>

<svelte:head>
	<title>Feeds</title>
</svelte:head>

<PageHead title="Feeds" className="justify-between">
	<Actions groups={data.groups} />
</PageHead>

<Tabs.Root value={currentGroup.toString()}>
	<Tabs.List>
		{#each data.groups.sort((a, b) => a.id - b.id) as g}
			<Tabs.Trigger value={g.id.toString()}>
				{#if g.feeds.find((f) => f.failure && !f.suspended) !== undefined}
					<AlertCircleIcon size="15" class="fill-destructive text-destructive-foreground mr-1" />
				{/if}
				{g.name}
			</Tabs.Trigger>
		{/each}
	</Tabs.List>
	{#each data.groups as g}
		{@const gf = [
			...g.feeds.filter((v) => v.failure && !v.suspended),
			...g.feeds.filter((v) => !v.failure && !v.suspended),
			...g.feeds.filter((v) => v.suspended)
		]}
		<Tabs.Content value={g.id.toString()}>
			<ul>
				{#each gf as f}
					<li>
						<Button
							class="flex items-center w-full h-12 py-2 px-4 text-start gap-2"
							variant="ghost"
							on:click={() => handleShowDetail(f)}
						>
							<span class="w-[18px]">
								{#if f.suspended}
									<CirclePauseIcon class="w-[18px]" />
								{:else if f.failure}
									<AlertCircleIcon class="w-[18px] fill-destructive text-destructive-foreground" />
								{/if}
							</span>
							<span class="inline-block w-1/2 truncate">{f.name}</span>
							<span class="inline-block w-1/2 truncate">{f.link}</span>
						</Button>
					</li>
				{/each}
			</ul>
		</Tabs.Content>
	{/each}
</Tabs.Root>

<Detail bind:show={showDetail} groups={data.groups} {selectedFeed} />
