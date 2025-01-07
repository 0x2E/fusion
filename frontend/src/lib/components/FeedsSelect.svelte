<script lang="ts">
	import type { Feed } from '$lib/api/model';
	import * as Command from '$lib/components/ui/command';
	import * as Popover from '$lib/components/ui/popover';
	import { cn } from '$lib/utils.js';
	import Check from 'lucide-svelte/icons/check';
	import ChevronsUpDown from 'lucide-svelte/icons/chevrons-up-down';
	import { tick } from 'svelte';
	import { Button, buttonVariants } from './ui/button';

	interface Props {
		data: Feed[];
		selected: number | undefined;
		onSelectedChange: (selected: number | undefined) => void;
		className?: string;
	}

	let { data, selected, onSelectedChange, className = '' }: Props = $props();
	let open = $state(false);

	let optionAll = { value: '-1', label: 'All' };
	let feeds = data
		.sort((a, b) => a.id - b.id)
		.map((f) => {
			return { value: String(f.id), label: f.name };
		});
	feeds.unshift(optionAll);

	let triggerRef = $state<HTMLButtonElement>(null!);

	// We want to refocus the trigger button when the user selects
	// an item from the list so users can continue navigating the
	// rest of the form with the keyboard.
	function closeAndFocusTrigger() {
		open = false;
		tick().then(() => {
			triggerRef.focus();
		});
	}
</script>

<Popover.Root bind:open>
	<Popover.Trigger
		bind:ref={triggerRef}
		class={cn(
			buttonVariants({ variant: 'outline' }),
			'w-[200px] justify-between overflow-hidden',
			className
		)}
	>
		{feeds.find((f) => f.value === String(selected))?.label || 'Select a feed...'}
		<ChevronsUpDown class="opacity-50" />
	</Popover.Trigger>
	<Popover.Content class="w-[200px] p-0">
		<Command.Root
			filter={(value, search) => {
				let name = '';
				if (value === optionAll.value) {
					name = optionAll.label;
				} else {
					name = data.find((v) => v.id === parseInt(value))?.name ?? '';
				}
				return name.toLowerCase().includes(search.toLowerCase()) ? 1 : 0;
			}}
		>
			<Command.Input placeholder="Search feed..." class="h-9" />
			<Command.List>
				<Command.Empty>No feed found.</Command.Empty>
				<Command.Group>
					{#each feeds as feed}
						<Command.Item
							value={feed.value}
							onSelect={() => {
								const id = parseInt(feed.value);
								if (id === -1) {
									onSelectedChange(undefined);
								} else {
									onSelectedChange(id);
								}
								closeAndFocusTrigger();
							}}
						>
							<Check class={cn(String(selected) !== feed.value && 'text-transparent')} />
							{feed.label}
						</Command.Item>
					{/each}
				</Command.Group>
			</Command.List>
		</Command.Root>
	</Popover.Content>
</Popover.Root>
