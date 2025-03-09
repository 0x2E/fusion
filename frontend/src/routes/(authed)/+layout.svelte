<script lang="ts">
	import Sidebar from '$lib/components/Sidebar.svelte';

	let { children, data } = $props();
</script>

<div class="drawer lg:drawer-open">
	<input id="sidebar-toggle" type="checkbox" class="drawer-toggle" />
	<div class="drawer-content relative overflow-x-clip">
		{@render children()}
	</div>
	<div class="drawer-side">
		<label for="sidebar-toggle" aria-label="close sidebar" class="drawer-overlay"></label>
		<div
			class="bg-base-200 text-base-content h-full min-h-full w-[80%] overflow-x-hidden px-2 py-4 lg:w-72"
		>
			<Sidebar feeds={data.feeds} />
		</div>
	</div>
</div>
