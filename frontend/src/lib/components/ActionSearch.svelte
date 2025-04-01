<script lang="ts">
	import { goto } from '$app/navigation';
	import { t } from '$lib/i18n';
	import { Search } from 'lucide-svelte';
</script>

<button
	onclick={async () => {
		await goto('/search');
	}}
	class="input input-sm lg:w-80"
>
	<Search class="size-4 opacity-50" />
	<input type="search" placeholder={t('item.search.placeholder')} readonly class="cursor-pointer" />
</button>
