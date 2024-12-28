<script lang="ts">
	import { goto } from '$app/navigation';
	import { login } from '$lib/api/login';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { toast } from 'svelte-sonner';

	let password = $state('');

	async function handleSubmit() {
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

<form onsubmit={handleSubmit} class="container max-w-[400px] mt-[20vh]">
	<h1 class="text-2xl font-bold text-center mt-10 mb-4">Login</h1>
	<div>
		<Label for="password">Password</Label>
		<Input name="password" type="password" autocomplete="current-password" bind:value={password} />
		<Button type="submit" class="w-full mt-4">Login</Button>
	</div>
</form>
