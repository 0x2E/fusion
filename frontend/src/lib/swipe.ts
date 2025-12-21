export interface SwipeCallbacks {
	onSwipeLeft?: () => void;
	onSwipeRight?: () => void;
}

interface SwipeOptions {
	minSwipeDistance?: number;
	maxVerticalDistance?: number;
}

export function swipe(
	node: HTMLElement,
	{ onSwipeLeft, onSwipeRight }: SwipeCallbacks,
	options: SwipeOptions = {}
) {
	const minSwipeDistance = options.minSwipeDistance ?? 50;
	const maxVerticalDistance = options.maxVerticalDistance ?? 100;

	let startX = 0;
	let startY = 0;
	let startTime = 0;

	function handleTouchStart(event: TouchEvent) {
		const touch = event.touches[0];
		startX = touch.clientX;
		startY = touch.clientY;
		startTime = Date.now();
	}

	function handleTouchEnd(event: TouchEvent) {
		const touch = event.changedTouches[0];
		const deltaX = touch.clientX - startX;
		const deltaY = touch.clientY - startY;
		const deltaTime = Date.now() - startTime;

		// Ignore if swipe is too slow (> 500ms) or too vertical
		if (deltaTime > 500 || Math.abs(deltaY) > maxVerticalDistance) {
			return;
		}

		const absX = Math.abs(deltaX);

		// Swipe right (next item)
		if (deltaX < -minSwipeDistance && absX > Math.abs(deltaY)) {
			onSwipeLeft?.();
		}
		// Swipe left (previous item)
		else if (deltaX > minSwipeDistance && absX > Math.abs(deltaY)) {
			onSwipeRight?.();
		}
	}

	node.addEventListener('touchstart', handleTouchStart, { passive: true });
	node.addEventListener('touchend', handleTouchEnd, { passive: true });

	return {
		update(newCallbacks: SwipeCallbacks) {
			onSwipeLeft = newCallbacks.onSwipeLeft;
			onSwipeRight = newCallbacks.onSwipeRight;
		},
		destroy() {
			node.removeEventListener('touchstart', handleTouchStart);
			node.removeEventListener('touchend', handleTouchEnd);
		}
	};
}
