<script lang="ts">
	interface Props {
		currentPage?: number;
		total: number;
		pageSize?: number;
		onPageChange: (page: number) => void;
	}

	let { currentPage = 1, total, pageSize = 10, onPageChange }: Props = $props();

	let totalPages = $derived(Math.ceil(total / pageSize));
	let pages = $derived(getPageNumbers(currentPage, totalPages));

	function getPageNumbers(current: number, total: number) {
		const pages: (number | string)[] = [];
		if (total <= 7) {
			return Array.from({ length: total }, (_, i) => i + 1);
		}

		pages.push(1);
		if (current > 3) {
			pages.push('...');
		}

		const start = Math.max(2, current - 1);
		const end = Math.min(total - 1, current + 1);

		for (let i = start; i <= end; i++) {
			pages.push(i);
		}

		if (current < total - 2) {
			pages.push('...');
		}
		pages.push(total);

		return pages;
	}

	function handlePageChange(page: number) {
		if (page < 1 || page > totalPages) return;
		onPageChange(page);
	}
</script>

<div class="join">
	<button
		class="join-item btn"
		disabled={currentPage === 1}
		onclick={() => handlePageChange(currentPage - 1)}>«</button
	>
	{#each pages as page}
		{#if typeof page === 'string'}
			<button class="join-item btn" disabled>...</button>
		{:else}
			<button
				class={`join-item btn ${page === currentPage ? 'btn-active border-b-base-content/60 border-b-2' : ''}`}
				onclick={() => handlePageChange(page)}
			>
				{page}
			</button>
		{/if}
	{/each}
	<button
		class="join-item btn"
		disabled={currentPage === totalPages}
		onclick={() => handlePageChange(currentPage + 1)}>»</button
	>
</div>
