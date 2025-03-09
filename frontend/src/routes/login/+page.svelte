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

<form
	onsubmit={handleSubmit}
	class="bg-base-200 container mx-auto mt-[20vh] flex max-w-[400px] flex-col rounded-md p-8"
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
