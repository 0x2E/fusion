import { create } from "zustand";

export interface ArticleListOverlay {
  readOverrides: Record<number, boolean>;
  starOverrides: Record<number, boolean>;
  stickyVisibleIds: Record<number, true>;
}

export function createEmptyArticleListOverlay(): ArticleListOverlay {
  return {
  readOverrides: {},
  starOverrides: {},
  stickyVisibleIds: {},
  };
}

interface ArticleSessionState {
  overlays: Record<string, ArticleListOverlay>;
  setReadOverride: (contextKey: string, itemId: number, unread: boolean) => void;
  setStarOverride: (contextKey: string, itemId: number, starred: boolean) => void;
  keepVisible: (contextKey: string, itemId: number) => void;
  setOverlay: (contextKey: string, overlay: ArticleListOverlay | undefined) => void;
  clearContext: (contextKey: string) => void;
}

export const useArticleSessionStore = create<ArticleSessionState>((set) => ({
  overlays: {},
  setReadOverride: (contextKey, itemId, unread) =>
    set((state) => ({
      overlays: {
        ...state.overlays,
        [contextKey]: {
          ...(state.overlays[contextKey] ?? createEmptyArticleListOverlay()),
          readOverrides: {
            ...(state.overlays[contextKey]?.readOverrides ?? {}),
            [itemId]: unread,
          },
        },
      },
    })),
  setStarOverride: (contextKey, itemId, starred) =>
    set((state) => ({
      overlays: {
        ...state.overlays,
        [contextKey]: {
          ...(state.overlays[contextKey] ?? createEmptyArticleListOverlay()),
          starOverrides: {
            ...(state.overlays[contextKey]?.starOverrides ?? {}),
            [itemId]: starred,
          },
        },
      },
    })),
  keepVisible: (contextKey, itemId) =>
    set((state) => ({
      overlays: {
        ...state.overlays,
        [contextKey]: {
          ...(state.overlays[contextKey] ?? createEmptyArticleListOverlay()),
          stickyVisibleIds: {
            ...(state.overlays[contextKey]?.stickyVisibleIds ?? {}),
            [itemId]: true,
          },
        },
      },
    })),
  setOverlay: (contextKey, overlay) =>
    set((state) => {
      const next = { ...state.overlays };

      if (overlay) {
        next[contextKey] = overlay;
      } else {
        delete next[contextKey];
      }

      return { overlays: next };
    }),
  clearContext: (contextKey) =>
    set((state) => {
      const next = { ...state.overlays };
      delete next[contextKey];
      return { overlays: next };
    }),
}));
