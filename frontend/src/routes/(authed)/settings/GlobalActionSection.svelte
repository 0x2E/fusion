<script lang="ts">
	import { listFeeds, refreshFeeds } from '$lib/api/feed';
	import { allGroups } from '$lib/api/group';
	import { dump } from '$lib/opml';
	import { toast } from 'svelte-sonner';
	import Section from './Section.svelte';

	async function handleRefreshAllFeeds() {
		if (!confirm('Are you sure you want to refresh all feeds except the suspended ones?')) {
			return;
		}
		try {
			await refreshFeeds({ all: true });
			toast.success('Start refreshing in the background');
		} catch (e) {
			toast.error((e as Error).message);
		}
	}

	async function handleExportAllFeeds() {
		// we don't use the gloabl state here because we need the latest data
		const groups = await allGroups();
		const feeds = await listFeeds();
		const data = groups.map((g) => {
			return {
				name: g.name,
				feeds: feeds
					.filter((f) => f.group.id === g.id)
					.map((f) => {
						return { name: f.name, link: f.link };
					})
			};
		});
		const content = dump(data);
		const link = document.createElement('a');
		link.href = 'data:text/xml;charset=utf-8,' + encodeURIComponent(content);
		link.download = 'feeds.opml';
		document.body.appendChild(link);
		link.click();
		document.body.removeChild(link);
	}
</script>

<Section id="global-actions" title="Global Actions">
	<div class="flex flex-wrap gap-2">
		<button onclick={() => handleRefreshAllFeeds()} class="btn btn-wide">Refresh all feeds</button>
		<button onclick={() => handleExportAllFeeds()} class="btn btn-wide">Export all feeds</button>
	</div>
</Section>
