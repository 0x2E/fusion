<script lang="ts">
	export let title: string;
	export let className = '';
</script>

<div class="flex items-center w-full {className}">
	<h1 class="text-3xl font-bold mb-4 ml-4">{title}</h1>
	<slot />
</div>
