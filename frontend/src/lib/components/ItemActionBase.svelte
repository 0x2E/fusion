<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import Button from './ui/button/button.svelte';
	import type { ComponentType } from 'svelte';
	import type { Icon } from 'lucide-svelte';
	import { cn } from '$lib/utils';

	interface Props {
		fn: (e: Event) => void;
		tooltip?: string;
		icon: ComponentType<Icon>;
		iconSize?: number;
		buttonClass?: string;
	}

	let {
		fn,
		tooltip = '',
		icon,
		iconSize = 18,
		buttonClass = ''
	}: Props = $props();
</script>

<Tooltip.Root openDelay={300}>
	<Tooltip.Trigger asChild >
		{#snippet children({ builder })}
				<Button
				builders={[builder]}
				variant="ghost"
				on:click={fn}
				class={cn('rounded-full', buttonClass)}
				size="icon"
				aria-label={tooltip}
			>
				{@const SvelteComponent = icon}
			<SvelteComponent size={iconSize} />
			</Button>
					{/snippet}
		</Tooltip.Trigger>
	<Tooltip.Content>
		<p>{tooltip}</p>
	</Tooltip.Content>
</Tooltip.Root>
