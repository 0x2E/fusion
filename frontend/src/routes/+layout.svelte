<script lang="ts">
	import '../app.css';
	import { Toaster } from '$lib/components/ui/sonner';
	import { ModeWatcher } from 'mode-watcher';
	import Footer from '$lib/components/Footer.svelte';
	import { page } from '$app/stores';
</script>

<svelte:head>
	<title>{$page.data.title ?? 'Fusion'}</title>
</svelte:head>

<ModeWatcher defaultMode="system" />
<Toaster position="bottom-right" richColors closeButton />

<!-- h-screen does not work properly on mobile. Use calc(100dvh) instead. 
     https://stackoverflow.com/a/76120728/12812480 -->
<main class="min-h-[calc(100dvh)] flex flex-col">
	<div class="flex flex-col grow max-w-[900px] w-full mx-auto">
		<slot />
	</div>
	<Footer />
</main>
