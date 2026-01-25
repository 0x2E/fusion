import { create } from "zustand";

export type ArticleFilter = "all" | "unread" | "starred";

interface UIState {
  // Navigation
  selectedGroupId: number | null;
  selectedFeedId: number | null;

  // Article drawer
  selectedArticleId: number | null;

  // Filters
  articleFilter: ArticleFilter;

  // Modals
  isSearchOpen: boolean;
  isSettingsOpen: boolean;

  // Actions
  setSelectedGroup: (groupId: number | null) => void;
  setSelectedFeed: (feedId: number | null) => void;
  setSelectedArticle: (articleId: number | null) => void;
  setArticleFilter: (filter: ArticleFilter) => void;
  setSearchOpen: (open: boolean) => void;
  setSettingsOpen: (open: boolean) => void;
  selectAll: () => void;
}

export const useUIStore = create<UIState>((set) => ({
  selectedGroupId: null,
  selectedFeedId: null,
  selectedArticleId: null,
  articleFilter: "all",
  isSearchOpen: false,
  isSettingsOpen: false,

  setSelectedGroup: (groupId) =>
    set({ selectedGroupId: groupId, selectedFeedId: null }),

  setSelectedFeed: (feedId) =>
    set({ selectedFeedId: feedId, selectedGroupId: null }),

  setSelectedArticle: (articleId) => set({ selectedArticleId: articleId }),

  setArticleFilter: (filter) => set({ articleFilter: filter }),

  setSearchOpen: (open) => set({ isSearchOpen: open }),

  setSettingsOpen: (open) => set({ isSettingsOpen: open }),

  selectAll: () => set({ selectedGroupId: null, selectedFeedId: null }),
}));
