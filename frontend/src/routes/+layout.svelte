<script lang="ts">
	import { page } from '$app/state';
	import { Toaster } from 'svelte-sonner';
	import '../app.css';

	let { children } = $props();
</script>

<svelte:head>
	<title>{page.data.title ?? 'Fusion'}</title>
</svelte:head>

<Toaster position="bottom-right" richColors />

<!-- h-screen does not work properly on mobile. Use calc(100dvh) instead. 
     https://stackoverflow.com/a/76120728/12812480 -->
<main class="min-h-[calc(100dvh)]">
	{@render children?.()}
</main>
