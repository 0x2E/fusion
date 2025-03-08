<script lang="ts">
	import Sidebar from '$lib/components/Sidebar.svelte';
	import type { LayoutProps } from './$types';

	let { children, data }: LayoutProps = $props();
</script>

<div class="drawer lg:drawer-open">
	<input id="sidebar-toggle" type="checkbox" class="drawer-toggle" />
	<div class="drawer-content">
		{@render children?.()}
	</div>
	<div class="drawer-side">
		<label for="sidebar-toggle" aria-label="close sidebar" class="drawer-overlay"></label>
		<div
			class="bg-base-200 text-base-content min-h-full h-full w-[80%] lg:w-72 py-4 px-2 overflow-x-hidden"
		>
			<Sidebar feeds={data.feeds} />
		</div>
	</div>
</div>
