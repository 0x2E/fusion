<script lang="ts">
	import { page } from '$app/stores';
</script>

<div class="text-center font-bold text-2xl">
	<h1>Error</h1>
	<h2>{$page.error?.message}</h2>
</div>
