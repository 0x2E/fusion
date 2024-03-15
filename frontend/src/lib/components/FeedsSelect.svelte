<script lang="ts">
	import Check from 'lucide-svelte/icons/check';
	import ChevronsUpDown from 'lucide-svelte/icons/chevrons-up-down';
	import * as Popover from '$lib/components/ui/popover';
	import * as Command from '$lib/components/ui/command';
	import { cn } from '$lib/utils.js';
	import { tick } from 'svelte';
	import type { Feed } from '$lib/api/model';
	import { Button } from './ui/button';

	export let data: Feed[];
	export let selected: number | undefined;
	export let className = '';
	let open = false;

	let optionAll = { value: '-1', label: 'All' };
	let feeds = data
		.sort((a, b) => a.id - b.id)
		.map((f) => {
			return { value: String(f.id), label: f.name };
		});
	feeds.unshift(optionAll);

	// We want to refocus the trigger button when the user selects
	// an item from the list so users can continue navigating the
	// rest of the form with the keyboard.
	function closeAndFocusTrigger(triggerId: string) {
		open = false;
		tick().then(() => {
			document.getElementById(triggerId)?.focus();
		});
	}
</script>

<Popover.Root bind:open let:ids>
	<Popover.Trigger asChild let:builder>
		<Button
			builders={[builder]}
			variant="outline"
			role="combobox"
			aria-expanded={open}
			class="w-[200px] justify-between {className}"
		>
			{feeds.find((f) => f.value === String(selected))?.label ?? 'Select a feed...'}
			<ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
		</Button>
	</Popover.Trigger>
	<Popover.Content class="w-[200px] p-0">
		<Command.Root
			filter={(value, search) => {
				// TODO: use better fuzz way: https://github.com/krisk/Fuse
				let name = '';
				if (value === optionAll.value) {
					name = optionAll.label;
				} else {
					name = data.find((v) => v.id === parseInt(value))?.name ?? '';
				}
				return name.includes(search) ? 1 : 0;
			}}
		>
			<Command.Input placeholder="Search feed..." />
			<Command.Empty>No feed found.</Command.Empty>
			<Command.Group>
				{#each feeds as f}
					<Command.Item
						value={String(f.value)}
						onSelect={(v) => {
							selected = parseInt(v);
							closeAndFocusTrigger(ids.trigger);
						}}
					>
						<Check class={cn('mr-2 h-4 w-4', String(selected) !== f.value && 'text-transparent')} />
						{f.label}
					</Command.Item>
				{/each}
			</Command.Group>
		</Command.Root>
	</Popover.Content>
</Popover.Root>
