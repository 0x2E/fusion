<script lang="ts">
	import { RefreshCw } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';

	interface Props {
		onRefresh: () => Promise<void>;
		children: Snippet;
		disabled?: boolean;
	}

	let { onRefresh, children, disabled = false }: Props = $props();

	let isRefreshing = $state(false);
	let isPulling = $state(false);
	let pullDistance = $state(0);
	let isIOSPWA = $state(false);

	const PULL_THRESHOLD = 40;
	const MAX_PULL_DISTANCE = 100;

	function detectIOSPWA(): boolean {
		const userAgent = navigator.userAgent;
		const isIOS = /iPad|iPhone|iPod/.test(userAgent) && !(window as any).MSStream;
		const isStandalone = window.matchMedia('(display-mode: standalone)').matches;
		const isWebApp = (navigator as any).standalone === true;
		const isInApp = /Safari/.test(userAgent) && /Version/.test(userAgent);

		return isIOS && (isStandalone || isWebApp || !isInApp);
	}

	let startY = 0;
	let currentY = 0;
	let touchStarted = false;
	let canPull = false;

	function handleTouchStart(event: TouchEvent) {
		if (disabled || isRefreshing) return;

		// Check if we're at the top of the page or container
		const scrollTop = document.documentElement.scrollTop || document.body.scrollTop;
		if (scrollTop > 5) return; // Allow for small scroll tolerance

		// Check if touch started near the top of the screen, accounting for safe area
		const touchY = event.touches[0].clientY;
		const safeAreaTop = window.screen?.height > 800 ? 44 : 20; // Estimate safe area for iPhone X+ vs older
		if (touchY > window.innerHeight * 0.3 + safeAreaTop) return; // Only allow pulls from top 30% of screen + safe area

		startY = touchY;
		touchStarted = true;
		canPull = true;
		isPulling = false;
		pullDistance = 0;
	}

	function handleTouchMove(event: TouchEvent) {
		if (!touchStarted || !canPull || disabled || isRefreshing) return;

		currentY = event.touches[0].clientY;
		const deltaY = currentY - startY;

		if (deltaY <= 0) {
			isPulling = false;
			pullDistance = 0;
			canPull = false;
			return;
		}

		const scrollTop = document.documentElement.scrollTop || document.body.scrollTop;
		if (scrollTop > 5) {
			isPulling = false;
			pullDistance = 0;
			canPull = false;
			return;
		}

		const resistance = Math.min(deltaY / 2.2, MAX_PULL_DISTANCE);
		pullDistance = resistance;

		if (resistance > 20) {
			isPulling = true;
			// Prevent page scroll when pulling with stronger resistance
			if (resistance > 40) {
				event.preventDefault();
			}
		}
	}

	async function handleTouchEnd(event: TouchEvent) {
		if (!touchStarted || disabled || isRefreshing) return;

		touchStarted = false;
		canPull = false;

		if (pullDistance > PULL_THRESHOLD) {
			isRefreshing = true;
			try {
				await onRefresh();
			} catch (error) {
				console.error('Pull to refresh error:', error);
			} finally {
				// Add a small delay for better UX
				setTimeout(() => {
					isRefreshing = false;
				}, 500);
			}
		}

		isPulling = false;
		pullDistance = 0;
	}

	function handleTouchCancel() {
		touchStarted = false;
		canPull = false;
		isPulling = false;
		pullDistance = 0;
	}

	onMount(() => {
		isIOSPWA = detectIOSPWA();

		if (isIOSPWA) {
			// Use passive listeners where possible for better performance
			document.addEventListener('touchstart', handleTouchStart, { passive: true });
			document.addEventListener('touchmove', handleTouchMove, { passive: false });
			document.addEventListener('touchend', handleTouchEnd, { passive: true });
			document.addEventListener('touchcancel', handleTouchCancel, { passive: true });

			return () => {
				document.removeEventListener('touchstart', handleTouchStart);
				document.removeEventListener('touchmove', handleTouchMove);
				document.removeEventListener('touchend', handleTouchEnd);
				document.removeEventListener('touchcancel', handleTouchCancel);
			};
		}
	});

	// Calculate the pull indicator state with smoother transitions
	let pullOpacity = $derived(Math.min(pullDistance / (PULL_THRESHOLD * 0.6), 1));
	let iconScale = $derived(Math.min(0.8 + (pullDistance / PULL_THRESHOLD) * 0.4, 1.2));
	let iconRotation = $derived(isRefreshing ? 0 : pullDistance * 1.8);
	let pullState = $derived(
		isRefreshing
			? 'refreshing'
			: pullDistance > PULL_THRESHOLD
				? 'ready'
				: isPulling
					? 'pulling'
					: 'idle'
	);
</script>

{#if isIOSPWA}
	<div class="relative">
		<!-- Pull to refresh indicator -->
		<div
			class="pointer-events-none fixed left-0 right-0 top-0 z-[60] flex items-end justify-center transition-all duration-200 ease-out"
			style="
				height: {Math.max(pullDistance + Math.max(window.screen?.height > 800 ? 44 : 20, 0), 0)}px; 
				opacity: {pullOpacity};
				background: linear-gradient(to bottom, oklch(var(--b1) / 0.95), oklch(var(--b1) / 0.8), transparent);
				padding-top: max(env(safe-area-inset-top, 0px), 20px);
			"
		>
			<div
				class="mb-2 flex flex-col items-center gap-1"
				style="padding-top: max(env(safe-area-inset-top, 0px), 10px);"
			>
				<div
					class="flex size-8 items-center justify-center rounded-full shadow-sm transition-all duration-300 ease-out"
					class:bg-info={pullState === 'ready' || pullState === 'refreshing'}
					class:text-info-content={pullState === 'ready' || pullState === 'refreshing'}
					class:bg-base-300={pullState === 'pulling'}
					class:text-base-content={pullState === 'pulling'}
					class:animate-spin={isRefreshing}
					style="
						transform: rotate({iconRotation}deg) scale({iconScale});
					"
				>
					<RefreshCw class="size-4 transition-all duration-300" />
				</div>
				<div
					class="text-xs font-medium transition-all duration-200"
					style="opacity: {pullOpacity * 0.9}; color: #6b7280;"
				>
					{#if isRefreshing}
						Refreshing...
					{:else if pullState === 'ready'}
						Release to refresh
					{:else if pullState === 'pulling'}
						Pull to refresh
					{/if}
				</div>
			</div>
		</div>

		<!-- Main content with transform applied during pull -->
		<div
			class="transition-transform duration-200 ease-out"
			style="transform: translateY({isPulling || isRefreshing
				? Math.min(pullDistance, MAX_PULL_DISTANCE)
				: 0}px)"
		>
			{@render children()}
		</div>
	</div>
{:else}
	{@render children()}
{/if}

<style>
	@keyframes spin {
		from {
			transform: rotate(0deg);
		}
		to {
			transform: rotate(360deg);
		}
	}

	.animate-spin {
		animation: spin 0.8s cubic-bezier(0.4, 0, 0.2, 1) infinite;
	}
</style>
