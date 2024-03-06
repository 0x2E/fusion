<script lang="ts">
	import '../app.css';
	import { Toaster } from '$lib/components/ui/sonner';
	import Navbar from '$lib/components/Navbar.svelte';
	import { ModeWatcher } from 'mode-watcher';
	import Footer from '$lib/components/Footer.svelte';
</script>

<svelte:head>
	<title>Fusion</title>
</svelte:head>

<ModeWatcher defaultMode="system" />
<Toaster position="top-right" richColors closeButton />
<main class="min-h-screen flex flex-col">
	<div class="flex flex-col grow max-w-[900px] w-full mx-auto">
		<Navbar />
		<div class="px-4 grow">
			<slot />
		</div>
	</div>
	<Footer />
</main>
