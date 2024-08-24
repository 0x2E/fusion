<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import Button from './ui/button/button.svelte';
	import type { ComponentType } from 'svelte';
	import type { Icon } from 'lucide-svelte';
	import { cn } from '$lib/utils';

	export let fn: (e: Event) => void;
	export let tooltip = '';
	export let icon: ComponentType<Icon>;
	export let iconSize = 18;
	export let buttonClass = '';
</script>

<Tooltip.Root openDelay={300}>
	<Tooltip.Trigger asChild let:builder>
		<Button
			builders={[builder]}
			variant="ghost"
			on:click={fn}
			class={cn('rounded-full', buttonClass)}
			size="icon"
			aria-label={tooltip}
		>
			<svelte:component this={icon} size={iconSize} />
		</Button>
	</Tooltip.Trigger>
	<Tooltip.Content>
		<p>{tooltip}</p>
	</Tooltip.Content>
</Tooltip.Root>
