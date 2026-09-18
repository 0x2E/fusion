import { create } from "zustand";
import type { ArticleFilter } from "@/lib/article-filter";

export type SidebarNode =
  | { key: string; kind: "filter"; filter: ArticleFilter }
  | { key: string; kind: "group"; groupId: number }
  | { key: string; kind: "feed"; feedId: number; groupId: number };

export const sidebarNodeKey = {
  filter: (filter: ArticleFilter) => `filter:${filter}`,
  group: (groupId: number) => `group:${groupId}`,
  feed: (feedId: number) => `feed:${feedId}`,
};

interface SidebarNavState {
  focusedKey: string | null;
  collapsedGroupIds: number[];
  setFocusedKey: (key: string | null) => void;
  setGroupCollapsed: (groupId: number, collapsed: boolean) => void;
  toggleGroupCollapsed: (groupId: number) => void;
}

export const useSidebarNavStore = create<SidebarNavState>((set) => ({
  focusedKey: null,
  collapsedGroupIds: [],

  setFocusedKey: (key) => set({ focusedKey: key }),

  setGroupCollapsed: (groupId, collapsed) =>
    set((state) => {
      const isCollapsed = state.collapsedGroupIds.includes(groupId);
      if (collapsed === isCollapsed) {
        return state;
      }
      return {
        collapsedGroupIds: collapsed
          ? [...state.collapsedGroupIds, groupId]
          : state.collapsedGroupIds.filter((id) => id !== groupId),
      };
    }),

  toggleGroupCollapsed: (groupId) =>
    set((state) => ({
      collapsedGroupIds: state.collapsedGroupIds.includes(groupId)
        ? state.collapsedGroupIds.filter((id) => id !== groupId)
        : [...state.collapsedGroupIds, groupId],
    })),
}));
