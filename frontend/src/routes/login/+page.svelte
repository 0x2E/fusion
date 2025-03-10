<script lang="ts">
	import { goto } from '$app/navigation';
	import { login } from '$lib/api/login';
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

<svelte:head>
	<title>Login</title>
</svelte:head>

<div class="flex h-[100vh] items-center justify-center">
	<form
		onsubmit={handleSubmit}
		class="bg-base-100 border-base-content/10 container flex max-w-[400px] -translate-y-[10vh] flex-col rounded-xl border p-8"
	>
		<h1 class="mb-4 text-center text-2xl font-bold">Login</h1>
		<fieldset class="fieldset">
			<legend class="fieldset-legend">Password</legend>
			<input
				name="password"
				type="password"
				autocomplete="current-password"
				bind:value={password}
				class="input w-full"
			/>
		</fieldset>
		<button type="submit" class="btn btn-primary mt-4 w-full">Login</button>
	</form>
</div>
