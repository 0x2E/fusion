<script module>
	let show = $state(false);
	export function toggleShow() {
		show = !show;
	}
</script>

<script lang="ts">
	import type { Component } from 'svelte';
	import FeedActionAddOne from './FeedActionAddOne.svelte';
	import FeedActionAddOpml from './FeedActionAddOPML.svelte';

	let modal = $state<HTMLDialogElement>();

	const tabs: { id: string; name: string; component: Component<any> }[] = [
		{ id: 'manually', name: 'Manually', component: FeedActionAddOne },
		{ id: 'import_opml', name: 'Import OPML', component: FeedActionAddOpml }
	];

	let selectedTabID = $state(tabs[0].id);
	let selectedTab = $derived(tabs.find((v) => v.id === selectedTabID) || tabs[0]);

	$effect(() => {
		if (show) {
			modal?.showModal();
		}
	});

	function doneCallback() {
		modal?.close();
	}
</script>

<dialog bind:this={modal} onclose={() => (show = false)} class="modal modal-bottom sm:modal-middle">
	<div class="modal-box">
		<form method="dialog">
			<button class="btn btn-sm btn-circle btn-ghost absolute top-2 right-2">✕</button>
		</form>
		<h3 class="text-lg font-bold">Add Feed(s)</h3>
		<div class="tabs tabs-box tabs-sm mt-2 w-fit">
			{#each tabs as tab}
				<input
					type="radio"
					name="add_feeds"
					class="tab text-xs font-medium"
					aria-label={tab.name}
					value={tab.id}
					bind:group={selectedTabID}
				/>
			{/each}
		</div>
		{#if show}
			<!-- used to destroy and recreate component to avoid resetting form manually -->
			<div>
				<selectedTab.component {doneCallback} />
			</div>
		{/if}
	</div>
	<form method="dialog" class="modal-backdrop">
		<button>close</button>
	</form>
</dialog>
