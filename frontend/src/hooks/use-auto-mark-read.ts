import { useEffect, useRef } from "react";

interface AutoMarkReadTarget {
  id: number;
  unread: boolean;
}

interface UseAutoMarkReadOptions {
  // Must stay non-null for the whole visit (while the drawer is open on a
  // markable article), regardless of unread flips or the feature being off:
  // gating it on either makes the id appear to change mid-visit, which
  // resets the one-shot latch below and reintroduces both the rollback
  // re-arm loop and the lost mark-unread veto.
  target: AutoMarkReadTarget | null;
  // Null disables the feature without ending the visit (unlike a null
  // target, it must not clear the veto).
  delayMs: number | null;
  onMarkRead: (id: number) => void;
  hasPendingMutation: () => boolean;
}

// A read/unread mutation typically settles well within this; the retry only
// needs to land on a render after isPending flipped back to false.
const PENDING_RETRY_MS = 500;

/**
 * One-shot auto mark-read per visit (#202).
 *
 * The countdown arms when `target` becomes a specific unread article id and
 * fires at most once for that visit. Deliberate constraints, each closing a
 * failure mode found in review:
 *
 * - The effect depends on the target id and delay only, never on `unread`
 *   or object identity: after firing (or after a failed mutation rolls the
 *   cache back to unread) the timer must not re-arm, or a network error
 *   turns delay 0 into a request loop.
 * - Observing the article read at any point during the visit vetoes firing
 *   for the rest of it, even if unread flips back (auto-mark rollback, or a
 *   manual read toggle undone before the deadline). The veto clears on a
 *   real leave only: the drawer closing, or moving to another article.
 * - "Still unread?" is read from refs at fire time, so a manual mark-read
 *   before the deadline makes firing a no-op.
 * - A pending read/unread mutation retries after a beat instead of
 *   consuming the visit: two overlapping snapshot/rollback mutations clobber
 *   each other's cache writes, and dropping the shot would let one slow
 *   request skip every article opened while it was in flight.
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
  const vetoedIdRef = useRef<number | null>(null);

  // Declared before the timer effect so a veto is already recorded or
  // cleared when the timer effect evaluates its arm condition on the same
  // render.
  useEffect(() => {
    if (target === null) {
      // Drawer closed: every future open is a fresh visit.
      vetoedIdRef.current = null;
      return;
    }
    if (vetoedIdRef.current !== null && vetoedIdRef.current !== target.id) {
      // Moving to another article is a leave for the vetoed one; without
      // this clear, next/previous navigation would stick the veto to its
      // article forever.
      vetoedIdRef.current = null;
    }
    if (!target.unread) {
      vetoedIdRef.current = target.id;
    }
  });

  const targetId = target?.id;

  useEffect(() => {
    if (
      targetId === undefined ||
      targetId === null ||
      delayMs === null ||
      !target?.unread ||
      vetoedIdRef.current === targetId
    ) {
      return;
    }
    const id = targetId;

    let timer: number | null = null;
    let fired = false;
    let remaining = delayMs;
    let deadline = Date.now() + remaining;

    const fire = () => {
      const current = targetRef.current;
      if (!current || current.id !== id || !current.unread || vetoedIdRef.current === id) {
        // Latch only when the callback runs or the shot is moot, never while
        // arming: a StrictMode remount must be able to clean up and re-arm.
        fired = true;
        return;
      }
      if (callbacksRef.current.hasPendingMutation()) {
        timer = window.setTimeout(onTimeout, PENDING_RETRY_MS);
        return;
      }
      fired = true;
      callbacksRef.current.onMarkRead(id);
    };

    const onTimeout = () => {
      timer = null;
      if (document.hidden) {
        // Do not fire in the background; the visibility handler resumes the
        // remainder when the document is seen again.
        remaining = Math.max(0, deadline - Date.now());
        return;
      }
      fire();
    };

    const schedule = (ms: number) => {
      timer = window.setTimeout(onTimeout, ms);
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
        schedule(remaining);
      }
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);
    // Arming while already hidden would burn the whole delay before the tab
    // is ever shown; wait for the first visible event instead.
    if (!document.hidden) {
      schedule(remaining);
    }

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
