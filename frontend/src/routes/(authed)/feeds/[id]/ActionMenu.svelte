<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { deleteFeed, updateFeed, type FeedUpdateForm } from '$lib/api/feed';
	import type { Feed } from '$lib/api/model';
	import { t } from '$lib/i18n';
	import { globalState } from '$lib/state.svelte';
	import { Ellipsis, Pause, Settings2, Trash } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		feed: Feed;
	}

	let { feed }: Props = $props();

	let settingsForm: FeedUpdateForm = $derived.by(() => {
		return {
			name: feed.name,
			link: feed.link,
			suspended: feed.suspended,
			req_proxy: feed.req_proxy,
			group_id: feed.group.id
		};
	});
	let settingsModal = $state<HTMLDialogElement>();

	const groups = $derived(globalState.groups);

	async function handleToggleSuspended() {
		try {
			await updateFeed(feed.id, {
				suspended: !feed.suspended
			});
			toast.success(t('state.success'));
		} catch (e) {
			toast.error((e as Error).message);
		}
		invalidateAll();
	}

	async function handleDelete() {
		if (!confirm(t('feed.delete.confirm'))) return;
		try {
			await deleteFeed(feed.id);
			toast.success(t('state.success'));
			await goto('/');
		} catch (e) {
			toast.error((e as Error).message);
		}
		invalidateAll();
	}

	async function handleUpdate(e: Event) {
		e.preventDefault();
		toast.promise(updateFeed(feed.id, settingsForm), {
			loading: 'Updating',
			success: () => {
				invalidateAll();
				settingsModal?.close();
				return t('state.success');
			},
			error: (e) => {
				invalidateAll();
				return (e as Error).message;
			}
		});
	}
</script>

<div class="dropdown dropdown-end">
	<div tabindex="0" role="button" class="btn btn-ghost btn-square">
		<Ellipsis class="size-4" />
	</div>
	<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
	<ul tabindex="0" class="dropdown-content menu bg-base-100 rounded-box z-1 w-52 p-2 shadow-sm">
		<li>
			<button
				onclick={() => {
					settingsModal?.showModal();
				}}
			>
				<Settings2 class="size-4" />
				<span> {t('common.settings')} </span>
			</button>
		</li>
		<li>
			<button onclick={handleToggleSuspended}>
				<Pause class="size-4" />
				<span>
					{feed.suspended ? t('feed.refresh.resume') : t('feed.refresh.suspend')}
				</span>
			</button>
		</li>
		<div class="divider my-0.5"></div>
		<li>
			<button onclick={handleDelete} class="text-error">
				<Trash class="size-4" />
				<span>{t('common.delete')}</span>
			</button>
		</li>
	</ul>
</div>

<dialog bind:this={settingsModal} class="modal modal-bottom sm:modal-middle">
	<div class="modal-box">
		<h3 class="text-lg font-bold">{t('common.settings')}</h3>
		<form class="w-full">
			<fieldset class="fieldset">
				<legend class="fieldset-legend">{t('common.name')}</legend>
				<input type="text" class="input w-full" bind:value={settingsForm.name} required />
			</fieldset>
			<fieldset class="fieldset">
				<legend class="fieldset-legend">{t('common.link')}</legend>
				<input type="url" class="input w-full" bind:value={settingsForm.link} required />
			</fieldset>
			<fieldset class="fieldset">
				<legend class="fieldset-legend">{t('common.group')}</legend>
				<select class="select" bind:value={settingsForm.group_id} required>
					{#each groups as group}
						<option value={group.id}>{group.name}</option>
					{/each}
				</select>
			</fieldset>

			<details class="mt-2">
				<summary>{t('common.advanced')}</summary>
				<div>
					<fieldset class="fieldset">
						<legend class="fieldset-legend">HTTP Proxy</legend>
						<input type="text" class="input w-full" bind:value={settingsForm.req_proxy} />
					</fieldset>
				</div>
			</details>
		</form>
		<div class="modal-action">
			<form method="dialog">
				<button class="btn btn-ghost">{t('common.cancel')}</button>
			</form>
			<button onclick={handleUpdate} class="btn btn-primary">{t('common.save')}</button>
		</div>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button>close</button>
	</form>
</dialog>
