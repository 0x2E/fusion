<script lang="ts">
	import { listFeeds, refreshFeeds } from '$lib/api/feed';
	import { allGroups } from '$lib/api/group';
	import { t } from '$lib/i18n';
	import { dump } from '$lib/opml';
	import { toast } from 'svelte-sonner';
	import Section from './Section.svelte';

	async function handleRefreshAllFeeds() {
		if (!confirm(t('feed.refresh.all.confirm'))) {
			return;
		}
		try {
			await refreshFeeds({ all: true });
			toast.success(t('feed.refresh.all.run_in_background'));
		} catch (e) {
			toast.error((e as Error).message);
		}
	}

	async function handleExportAllFeeds() {
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

<Section id="global-actions" title={t('settings.global_actions')}>
	<div class="flex flex-wrap gap-2">
		<button onclick={() => handleRefreshAllFeeds()} class="btn btn-wide"
			>{t('settings.global_actions.refresh_all_feeds')}</button
		>
		<button onclick={() => handleExportAllFeeds()} class="btn btn-wide"
			>{t('settings.global_actions.export_all_feeds')}</button
		>
	</div>
</Section>
