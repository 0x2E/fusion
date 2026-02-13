import { create } from "zustand";

interface ArticleSessionState {
  starredOverrides: Record<number, boolean>;
  setStarredOverride: (itemId: number, starred: boolean) => void;
  clearStarredOverride: (itemId: number) => void;
}

export const useArticleSessionStore = create<ArticleSessionState>((set) => ({
  starredOverrides: {},
  setStarredOverride: (itemId, starred) =>
    set((state) => ({
      starredOverrides: {
        ...state.starredOverrides,
        [itemId]: starred,
      },
    })),
  clearStarredOverride: (itemId) =>
    set((state) => {
      const next = { ...state.starredOverrides };
      delete next[itemId];
      return { starredOverrides: next };
    }),
}));
