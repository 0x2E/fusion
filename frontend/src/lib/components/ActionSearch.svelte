<script lang="ts">
	let isMac = navigator.platform.indexOf('Mac') === 0 || navigator.platform === 'iPhone';

	async function handleSearch() {
		//
	}
</script>

<label class="input input-sm lg:w-80">
	<svg class="size-4 opacity-50" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"
		><g
			stroke-linejoin="round"
			stroke-linecap="round"
			stroke-width="2.5"
			fill="none"
			stroke="currentColor"
			><circle cx="11" cy="11" r="8"></circle><path d="m21 21-4.3-4.3"></path></g
		></svg
	>

	<input type="search" class="grow" placeholder="Search in title and content" />
	<!-- <kbd class="kbd kbd-sm">{isMac ? '⌘' : '^'}</kbd>
	<kbd class="kbd kbd-sm">K</kbd> -->
</label>
