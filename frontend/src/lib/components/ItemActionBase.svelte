<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { cn } from '$lib/utils';
	import type { Icon as IconType } from 'lucide-svelte';
	import { buttonVariants } from './ui/button';

	interface Props {
		fn: (e: Event) => void;
		tooltip?: string;
		icon: typeof IconType;
		iconSize?: number;
		buttonClass?: string;
	}

	let { fn, tooltip = '', icon: Icon, iconSize = 18, buttonClass = '' }: Props = $props();
</script>

<Tooltip.Provider>
	<Tooltip.Root delayDuration={100}>
		<Tooltip.Trigger
			onclick={fn}
			aria-label={tooltip}
			class={cn(buttonVariants({ variant: 'ghost', size: 'icon' }), 'rounded-full', buttonClass)}
		>
			<Icon size={iconSize} />
		</Tooltip.Trigger>
		<Tooltip.Content>
			<p>{tooltip}</p>
		</Tooltip.Content>
	</Tooltip.Root>
</Tooltip.Provider>
