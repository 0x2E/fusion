import { useEffect, useRef } from "react";

interface AutoMarkReadTarget {
  id: number;
  unread: boolean;
}

interface UseAutoMarkReadOptions {
  // Null (feature off, drawer closed, article not unread) disarms the timer.
  // Unread is read once, when the target arms the countdown: a manual
  // mark-unread afterwards is a "keep unread" veto and must not re-arm.
  target: AutoMarkReadTarget | null;
  delayMs: number;
  onMarkRead: (id: number) => void;
  hasPendingMutation: () => boolean;
}

/**
 * One-shot auto mark-read per visit (#202).
 *
 * The countdown arms when `target` becomes a specific unread article id and
 * fires at most once for that visit. Deliberate constraints, each closing a
 * failure mode found in review:
 *
 * - The effect depends on the target id and delay only, never on
 *   `unread`: after firing (or after a failed mutation rolls the cache back
 *   to unread) the timer must not re-arm, or a network error turns delay 0
 *   into a request loop.
 * - "Still unread?" is read from a ref at fire time, so a manual mark-read
 *   before the deadline makes firing a no-op.
 * - `onMarkRead` must land `mutate` and `pinRead` in the same synchronous
 *   turn, or the unread list drops the row before the pin keeps it visible.
 * - Time spent with the document hidden does not count: a phone "misclick"
 *   gesture is leaving the app, and a background tab's timers fire in a
 *   burst on return.
 */
export function useAutoMarkRead({
  target,
  delayMs,
  onMarkRead,
  hasPendingMutation,
}: UseAutoMarkReadOptions) {
  // Latest target and callbacks for the fire-time checks; keeping them in
  // refs avoids re-arming on object identity changes (detail refetches
  // replace the article object without changing the id).
  const targetRef = useRef(target);
  targetRef.current = target;
  const callbacksRef = useRef({ onMarkRead, hasPendingMutation });
  callbacksRef.current = { onMarkRead, hasPendingMutation };

  const targetId = target?.id;

  useEffect(() => {
    if (targetId === undefined || targetId === null || !target?.unread) {
      return;
    }
    const id = targetId;

    let timer: number | null = null;
    let remaining = delayMs;
    let deadline = Date.now() + remaining;
    let fired = false;

    const fire = () => {
      // Latch when the callback runs, not when the effect arms, so a
      // StrictMode remount can clean up and re-arm freely in dev.
      fired = true;
      const current = targetRef.current;
      if (!current || current.id !== id || !current.unread) {
        return;
      }
      if (callbacksRef.current.hasPendingMutation()) {
        return;
      }
      callbacksRef.current.onMarkRead(id);
    };

    const schedule = () => {
      timer = window.setTimeout(() => {
        timer = null;
        if (document.hidden) {
          // Do not fire in the background; the visibility handler resumes
          // the remainder when the document is seen again.
          remaining = Math.max(0, deadline - Date.now());
          return;
        }
        fire();
      }, remaining);
    };

    const handleVisibilityChange = () => {
      if (fired) {
        return;
      }
      if (document.hidden) {
        if (timer !== null) {
          window.clearTimeout(timer);
          timer = null;
          remaining = Math.max(0, deadline - Date.now());
        }
      } else if (timer === null) {
        deadline = Date.now() + remaining;
        schedule();
      }
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);
    schedule();

    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange);
      if (timer !== null) {
        window.clearTimeout(timer);
      }
    };
    // target?.unread is intentionally absent: see the hook docs.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [targetId, delayMs]);
}
