<script lang="ts">
	import { Pagination as PaginationPrimitive } from "bits-ui";
	import ChevronLeft from "lucide-svelte/icons/chevron-left";
	import { cn } from "$lib/utils.js";
	import { buttonVariants } from "$lib/components/ui/button/index.js";

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: PaginationPrimitive.PrevButtonProps = $props();
</script>

{#snippet Fallback()}
	<ChevronLeft />
	<span class="hidden sm:block">Previous</span>
{/snippet}

<PaginationPrimitive.PrevButton
	bind:ref
	{...restProps}
	class={cn(
		buttonVariants({ variant: 'ghost', className: 'gap-1 px-0 pl-0 md:px-4 md:pl-2.5' }),
		className
	)}
	children={children || Fallback}
/>
