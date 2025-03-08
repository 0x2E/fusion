<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		title: string;
		className?: string;
		children?: Snippet;
	}

	let { title, className = '', children }: Props = $props();
</script>

<div class="flex items-center w-full {className}">
	<h1 class="text-3xl font-bold mb-4 ml-4">{title}</h1>
	{@render children?.()}
</div>
