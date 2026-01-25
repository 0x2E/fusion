import { create } from "zustand";
import type { Group, Feed, Item, Bookmark } from "@/lib/api";

interface DataState {
  // Data
  groups: Group[];
  feeds: Feed[];
  items: Item[];
  bookmarks: Bookmark[];

  // Loading states
  isLoadingGroups: boolean;
  isLoadingFeeds: boolean;
  isLoadingItems: boolean;
  isLoadingBookmarks: boolean;

  // Error states
  groupsError: string | null;
  feedsError: string | null;
  itemsError: string | null;
  bookmarksError: string | null;

  // Actions
  setGroups: (groups: Group[]) => void;
  setFeeds: (feeds: Feed[]) => void;
  setItems: (items: Item[]) => void;
  setBookmarks: (bookmarks: Bookmark[]) => void;

  setLoadingGroups: (loading: boolean) => void;
  setLoadingFeeds: (loading: boolean) => void;
  setLoadingItems: (loading: boolean) => void;
  setLoadingBookmarks: (loading: boolean) => void;

  setGroupsError: (error: string | null) => void;
  setFeedsError: (error: string | null) => void;
  setItemsError: (error: string | null) => void;
  setBookmarksError: (error: string | null) => void;

  // Item mutations
  markItemRead: (itemId: number) => void;
  markItemUnread: (itemId: number) => void;
  markItemsRead: (itemIds: number[]) => void;

  // Bookmark mutations
  addBookmark: (bookmark: Bookmark) => void;
  removeBookmark: (bookmarkId: number) => void;

  // Helpers
  getFeedById: (feedId: number) => Feed | undefined;
  getGroupById: (groupId: number) => Group | undefined;
  getFeedsByGroup: (groupId: number) => Feed[];
  getItemById: (itemId: number) => Item | undefined;
  isItemStarred: (itemId: number) => boolean;
  getBookmarkByItemId: (itemId: number) => Bookmark | undefined;
}

export const useDataStore = create<DataState>((set, get) => ({
  groups: [],
  feeds: [],
  items: [],
  bookmarks: [],

  isLoadingGroups: false,
  isLoadingFeeds: false,
  isLoadingItems: false,
  isLoadingBookmarks: false,

  groupsError: null,
  feedsError: null,
  itemsError: null,
  bookmarksError: null,

  setGroups: (groups) => set({ groups }),
  setFeeds: (feeds) => set({ feeds }),
  setItems: (items) => set({ items }),
  setBookmarks: (bookmarks) => set({ bookmarks }),

  setLoadingGroups: (loading) => set({ isLoadingGroups: loading }),
  setLoadingFeeds: (loading) => set({ isLoadingFeeds: loading }),
  setLoadingItems: (loading) => set({ isLoadingItems: loading }),
  setLoadingBookmarks: (loading) => set({ isLoadingBookmarks: loading }),

  setGroupsError: (error) => set({ groupsError: error }),
  setFeedsError: (error) => set({ feedsError: error }),
  setItemsError: (error) => set({ itemsError: error }),
  setBookmarksError: (error) => set({ bookmarksError: error }),

  markItemRead: (itemId) =>
    set((state) => ({
      items: state.items.map((item) =>
        item.id === itemId ? { ...item, unread: false } : item
      ),
    })),

  markItemUnread: (itemId) =>
    set((state) => ({
      items: state.items.map((item) =>
        item.id === itemId ? { ...item, unread: true } : item
      ),
    })),

  markItemsRead: (itemIds) =>
    set((state) => ({
      items: state.items.map((item) =>
        itemIds.includes(item.id) ? { ...item, unread: false } : item
      ),
    })),

  addBookmark: (bookmark) =>
    set((state) => ({ bookmarks: [bookmark, ...state.bookmarks] })),

  removeBookmark: (bookmarkId) =>
    set((state) => ({
      bookmarks: state.bookmarks.filter((b) => b.id !== bookmarkId),
    })),

  getFeedById: (feedId) => get().feeds.find((f) => f.id === feedId),

  getGroupById: (groupId) => get().groups.find((g) => g.id === groupId),

  getFeedsByGroup: (groupId) =>
    get().feeds.filter((f) => f.group_id === groupId),

  getItemById: (itemId) => get().items.find((i) => i.id === itemId),

  isItemStarred: (itemId) =>
    get().bookmarks.some((b) => b.item_id === itemId),

  getBookmarkByItemId: (itemId) =>
    get().bookmarks.find((b) => b.item_id === itemId),
}));
