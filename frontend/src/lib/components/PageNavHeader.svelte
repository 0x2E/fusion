<script lang="ts">
	import { Menu } from 'lucide-svelte';
	import type { Snippet } from 'svelte';

	interface Props {
		children?: Snippet;
		title?: string;
	}

	let { title, children }: Props = $props();
</script>

<header class="sticky top-0 bg-base-100 border-b border-base-content/10 py-2 mb-8 z-50">
	<div class="flex justify-between items-center px-4">
		<div>
			<label for="sidebar-toggle" class="btn btn-ghost btn-square lg:hidden">
				<Menu class="size-5" />
			</label>
			{#if title}
				<span class="text-sm text-base-content/60">{title}</span>
			{/if}
		</div>
		<div>
			{@render children?.()}
		</div>
	</div>
</header>
