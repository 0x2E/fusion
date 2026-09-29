import { useCallback } from "react";
import { create } from "zustand";

/**
 * Read-state pins keep items marked read visible (grayed) in the unread view
 * until the user leaves that exact list, so the mark can still be undone in
 * place instead of hunting the item down in the all view. Pins are pure view
 * state: they never touch query caches and are never persisted — leaving the
 * list identity (switching filter/feed/group or unmounting the view) drops
 * them, which is the moment the read marking is considered committed.
 */

export function articleListKey(
  feedId: number | null,
  groupId: number | null,
  filter: string,
): string {
  return `${feedId ?? 0}:${groupId ?? 0}:${filter}`;
}

interface ArticlePinsState {
  pins: Record<string, Set<number>>;
  pinItems: (key: string, ids: number[]) => void;
  unpinItems: (key: string, ids: number[]) => void;
  clearPins: (key: string) => void;
}

export const useArticlePinsStore = create<ArticlePinsState>()((set) => ({
  pins: {},
  pinItems: (key, ids) =>
    set((state) => {
      if (ids.length === 0) return state;
      const next = new Set(state.pins[key]);
      let added = false;
      for (const id of ids) {
        if (id > 0 && !next.has(id)) {
          next.add(id);
          added = true;
        }
      }
      if (!added) return state;
      return { pins: { ...state.pins, [key]: next } };
    }),
  unpinItems: (key, ids) =>
    set((state) => {
      const current = state.pins[key];
      if (!current || ids.length === 0) return state;
      const next = new Set(current);
      for (const id of ids) {
        next.delete(id);
      }
      return { pins: { ...state.pins, [key]: next } };
    }),
  clearPins: (key) =>
    set((state) => {
      if (!(key in state.pins)) return state;
      const pins = { ...state.pins };
      delete pins[key];
      return { pins };
    }),
}));

/**
 * Write side of the pin mechanism. Pinning only applies to the unread view:
 * items marked read elsewhere (e.g. from the all view) are committed
 * immediately — otherwise switching back to unread would keep showing them,
 * which is the very behavior issue #262 reports.
 */
export function useReadStatePins(filters: {
  feedId: number | null;
  groupId: number | null;
  articleFilter: string;
}) {
  const pinItems = useArticlePinsStore((s) => s.pinItems);
  const unpinItems = useArticlePinsStore((s) => s.unpinItems);
  const key = articleListKey(
    filters.feedId,
    filters.groupId,
    filters.articleFilter,
  );
  const active = filters.articleFilter === "unread";

  const pinRead = useCallback(
    (ids: number[]) => {
      if (active) pinItems(key, ids);
    },
    [active, key, pinItems],
  );

  const unpinRead = useCallback(
    (ids: number[]) => {
      if (active) unpinItems(key, ids);
    },
    [active, key, unpinItems],
  );

  return { pinRead, unpinRead };
}
