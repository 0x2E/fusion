import { useCallback, useRef } from "react";

interface SwipeOptions {
  /** Minimum horizontal distance (px) to trigger a swipe. Default: 50 */
  threshold?: number;
  /**
   * If set, only trigger when the touch starts within this many pixels
   * from the left edge of the screen. Useful for sidebar open gestures.
   */
  fromLeftEdge?: number;
}

/**
 * Detects horizontal swipe gestures and calls the appropriate callback.
 * Vertical scrolling is unaffected — only fires when horizontal movement
 * dominates and exceeds the threshold.
 */
export function useSwipeNavigation(
  onSwipeLeft: () => void,
  onSwipeRight: () => void,
  options: SwipeOptions = {},
) {
  const { threshold = 50, fromLeftEdge } = options;
  const startX = useRef(0);
  const startY = useRef(0);

  const onTouchStart = useCallback((e: React.TouchEvent) => {
    startX.current = e.touches[0].clientX;
    startY.current = e.touches[0].clientY;
  }, []);

  const onTouchEnd = useCallback(
    (e: React.TouchEvent) => {
      if (fromLeftEdge !== undefined && startX.current > fromLeftEdge) return;
      const dx = e.changedTouches[0].clientX - startX.current;
      const dy = e.changedTouches[0].clientY - startY.current;
      if (Math.abs(dx) < threshold || Math.abs(dx) < Math.abs(dy)) return;
      if (dx < 0) onSwipeLeft();
      else onSwipeRight();
    },
    [onSwipeLeft, onSwipeRight, threshold, fromLeftEdge],
  );

  return { onTouchStart, onTouchEnd };
}
