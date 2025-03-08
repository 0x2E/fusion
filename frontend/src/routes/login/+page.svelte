<script lang="ts">
	import { goto } from '$app/navigation';
	import { login } from '$lib/api/login';
	import { toast } from 'svelte-sonner';

	let password = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();

		try {
			await login(password);
			goto('/');
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
	class="container flex flex-col bg-base-200 mx-auto max-w-[400px] mt-[20vh] p-8 rounded-md"
>
	<h1 class="text-2xl font-bold text-center mb-4">Login</h1>
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
	<button type="submit" class="btn btn-primary w-full mt-4">Login</button>
</form>
