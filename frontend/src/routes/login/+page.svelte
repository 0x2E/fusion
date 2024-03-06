<script lang="ts">
	import { goto } from '$app/navigation';
	import { login } from '$lib/api/login';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { toast } from 'svelte-sonner';

	let password = '';

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

<form on:submit={handleSubmit} class="max-w-[600px] mx-auto">
	<h1 class="text-2xl font-bold text-center mt-10 mb-4">Login</h1>
	<div class="space-y-2">
		<Label for="password">Password</Label>
		<Input name="password" bind:value={password} />
		<Button type="submit" class="w-full">Login</Button>
	</div>
</form>
