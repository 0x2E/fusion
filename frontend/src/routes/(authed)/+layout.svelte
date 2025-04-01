<script lang="ts">
	import { beforeNavigate } from '$app/navigation';
	import FeedActionImport from '$lib/components/FeedActionImport.svelte';
	import ShortcutHelpModal from '$lib/components/ShortcutHelpModal.svelte';
	import Sidebar from '$lib/components/Sidebar.svelte';

	let { children, data } = $props();
	let showSidebar = $state(false);
	beforeNavigate(() => {
		showSidebar = false;
	});
</script>

<div class="drawer lg:drawer-open">
	<input id="sidebar-toggle" type="checkbox" bind:checked={showSidebar} class="drawer-toggle" />
	<div class="drawer-content bg-base-100 relative z-10 min-h-screen overflow-x-clip">
		<div class="mx-auto flex h-full max-w-6xl flex-col pb-4">
			<svelte:boundary>
				{@render children()}
				{#snippet failed(error, reset)}
					<p>{error}</p>
					<button onclick={reset} class="btn w-fit">oops! try again</button>
				{/snippet}
			</svelte:boundary>
		</div>
	</div>
	<div class="drawer-side z-10">
		<label for="sidebar-toggle" aria-label="close sidebar" class="drawer-overlay"></label>
		<div
			class="text-base-content bg-base-200 z-50 h-full min-h-full w-[80%] overflow-x-hidden px-2 py-4 lg:w-72"
		>
			<Sidebar feeds={data.feeds} groups={data.groups} />
		</div>
	</div>
</div>

<!-- put these outside the drawer because when its inner modal is placed inside the drawer sidebar, the underlying dialog won't close properly -->
<FeedActionImport />
<ShortcutHelpModal />
