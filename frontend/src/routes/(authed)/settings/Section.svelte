<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		id: string;
		title: string;
		description?: string;
		children: Snippet;
	}
	let { id, title, description, children }: Props = $props();
</script>

<section {id} class="w-full">
	<div class="border-base-300 rounded-xl border p-4 shadow">
		<h2 class="font-semibold">{title}</h2>
		{#if description}
			<p class="text-base-content/60 text-xs">{description}</p>
		{/if}
		<div class="mt-4">
			{@render children()}
		</div>
	</div>
</section>
