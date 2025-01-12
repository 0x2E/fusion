<script lang="ts">
	import Navbar from '$lib/components/Navbar.svelte';
	interface Props {
		children?: import('svelte').Snippet;
	}

	let { children }: Props = $props();
</script>

<Navbar />
<div class="px-4 grow">
	{@render children?.()}
</div>
