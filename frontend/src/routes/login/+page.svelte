<script lang="ts">
	import { goto } from '$app/navigation';
	import { login } from '$lib/api/login';
	import { t } from '$lib/i18n';
	import { toast } from 'svelte-sonner';

	let password = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();

		try {
			await login(password);
			await goto('/');
		} catch (e) {
			toast.error((e as Error).message);
		}
	}
</script>

<div class="flex h-[100vh] items-center justify-center">
	<form
		onsubmit={handleSubmit}
		class="border-base-content/10 container flex max-w-[400px] -translate-y-[10vh] flex-col rounded-xl border p-8 shadow"
	>
		<h1 class="mb-4 text-center text-2xl font-bold">Fusion</h1>
		<fieldset class="fieldset">
			<legend class="fieldset-legend">{t('common.password')}</legend>
			<input
				name="password"
				type="password"
				autocomplete="current-password"
				bind:value={password}
				class="input w-full"
			/>
		</fieldset>
		<button type="submit" class="btn btn-primary mt-4 w-full">{t('common.login')}</button>
	</form>
</div>
