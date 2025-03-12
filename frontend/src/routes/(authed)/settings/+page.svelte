<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import PageNavHeader from '$lib/components/PageNavHeader.svelte';
	import { onMount } from 'svelte';
	import GlobalActionSection from './GlobalActionSection.svelte';
	import GroupSection from './GroupSection.svelte';

	const links: {
		label: string;
		hash: string;
	}[] = [
		{ label: 'Global Actions', hash: '#global-actions' },
		{ label: 'Groups', hash: '#groups' }
	];

	onMount(() => {
		const url = page.url;
		if (!links.map((v) => v.hash).includes(url.hash)) {
			url.hash = links[0].hash;
			goto(url);
		}
	});
</script>

<svelte:head>
	<title>Settings</title>
</svelte:head>

<div class="flex flex-col">
	<PageNavHeader title="Settings"></PageNavHeader>
	<div class="px-4 lg:px-8">
		<div class="py-6">
			<h1 class="text-3xl font-bold">Settings</h1>
		</div>
		<div class="relative flex flex-col gap-6 lg:flex-row lg:gap-14">
			<div class="w-full lg:w-52">
				<ul class="menu rounded-box sticky top-10 w-full">
					{#each links as link}
						<li>
							<a
								href={link.hash}
								class="font-medium"
								class:menu-active={page.url.hash === link.hash}
							>
								{link.label}
							</a>
						</li>
					{/each}
				</ul>
			</div>
			<div class="flex grow flex-col gap-6">
				<GlobalActionSection />
				<GroupSection />
			</div>
		</div>
	</div>
</div>
