<script lang="ts">
	import Sidebar from '$lib/components/Sidebar.svelte';
	import { Menu } from 'lucide-svelte';
	import type { LayoutProps } from './$types';
	import ThemeController from '$lib/components/ThemeController.svelte';

	let { children, data }: LayoutProps = $props();
</script>

<div class="drawer lg:drawer-open">
	<input id="sidebar-toggle" type="checkbox" class="drawer-toggle" />
	<div class="drawer-content">
		<div class="flex flex-col gap-4 px-2">
			<header
				class="flex items-center sticky top-0 bg-base-100 border-b border-base-content/10 py-2"
			>
				<label for="sidebar-toggle" class="btn btn-ghost btn-square lg:hidden">
					<Menu class="size-5" />
				</label>
			</header>
			<div class="px-4">
				{@render children?.()}
			</div>
		</div>
	</div>
	<div class="drawer-side">
		<label for="sidebar-toggle" aria-label="close sidebar" class="drawer-overlay"></label>
		<div
			class="bg-base-200 text-base-content min-h-full w-[80%] lg:w-72 py-4 px-2 overflow-x-hidden"
		>
			<Sidebar feeds={data.feeds} />
		</div>
	</div>
</div>
