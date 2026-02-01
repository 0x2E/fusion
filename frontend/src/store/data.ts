import { create } from "zustand";
import type { Group, Feed, Item, Bookmark } from "@/lib/api";

interface DataState {
  // Data
  groups: Group[];
  feeds: Feed[];
  items: Item[];
  itemsTotal: number;
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
  appendItems: (items: Item[]) => void;
  setItemsTotal: (total: number) => void;
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

  // Feed unread count mutations
  decrementFeedUnreadCount: (feedId: number) => void;
  incrementFeedUnreadCount: (feedId: number) => void;
  decrementFeedUnreadCounts: (feedCounts: Map<number, number>) => void;

  // Bookmark mutations
  addBookmark: (bookmark: Bookmark) => void;
  removeBookmark: (bookmarkId: number) => void;

  // Group mutations
  addGroup: (group: Group) => void;
  updateGroup: (groupId: number, name: string) => void;
  removeGroup: (groupId: number) => void;

  // Feed mutations
  addFeed: (feed: Feed) => void;
  updateFeed: (feedId: number, updates: Partial<Feed>) => void;
  removeFeed: (feedId: number) => void;
  updateFeedGroup: (feedId: number, groupId: number) => void;
  moveFeedsToGroup: (fromGroupId: number, toGroupId: number) => void;

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
  itemsTotal: 0,
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
  appendItems: (items) =>
    set((state) => ({ items: [...state.items, ...items] })),
  setItemsTotal: (total) => set({ itemsTotal: total }),
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
    set((state) => {
      const item = state.items.find((i) => i.id === itemId);
      if (!item || !item.unread) return state;
      return {
        items: state.items.map((i) =>
          i.id === itemId ? { ...i, unread: false } : i
        ),
        feeds: state.feeds.map((f) =>
          f.id === item.feed_id
            ? { ...f, unread_count: Math.max(0, f.unread_count - 1) }
            : f
        ),
      };
    }),

  markItemUnread: (itemId) =>
    set((state) => {
      const item = state.items.find((i) => i.id === itemId);
      if (!item || item.unread) return state;
      return {
        items: state.items.map((i) =>
          i.id === itemId ? { ...i, unread: true } : i
        ),
        feeds: state.feeds.map((f) =>
          f.id === item.feed_id ? { ...f, unread_count: f.unread_count + 1 } : f
        ),
      };
    }),

  markItemsRead: (itemIds) =>
    set((state) => {
      const feedCounts = new Map<number, number>();
      for (const item of state.items) {
        if (itemIds.includes(item.id) && item.unread) {
          feedCounts.set(item.feed_id, (feedCounts.get(item.feed_id) || 0) + 1);
        }
      }
      return {
        items: state.items.map((item) =>
          itemIds.includes(item.id) ? { ...item, unread: false } : item
        ),
        feeds: state.feeds.map((f) => {
          const count = feedCounts.get(f.id);
          return count
            ? { ...f, unread_count: Math.max(0, f.unread_count - count) }
            : f;
        }),
      };
    }),

  decrementFeedUnreadCount: (feedId) =>
    set((state) => ({
      feeds: state.feeds.map((f) =>
        f.id === feedId
          ? { ...f, unread_count: Math.max(0, f.unread_count - 1) }
          : f
      ),
    })),

  incrementFeedUnreadCount: (feedId) =>
    set((state) => ({
      feeds: state.feeds.map((f) =>
        f.id === feedId ? { ...f, unread_count: f.unread_count + 1 } : f
      ),
    })),

  decrementFeedUnreadCounts: (feedCounts) =>
    set((state) => ({
      feeds: state.feeds.map((f) => {
        const count = feedCounts.get(f.id);
        return count
          ? { ...f, unread_count: Math.max(0, f.unread_count - count) }
          : f;
      }),
    })),

  addBookmark: (bookmark) =>
    set((state) => ({ bookmarks: [bookmark, ...state.bookmarks] })),

  removeBookmark: (bookmarkId) =>
    set((state) => ({
      bookmarks: state.bookmarks.filter((b) => b.id !== bookmarkId),
    })),

  addGroup: (group) =>
    set((state) => ({ groups: [...state.groups, group] })),

  updateGroup: (groupId, name) =>
    set((state) => ({
      groups: state.groups.map((g) =>
        g.id === groupId ? { ...g, name } : g
      ),
    })),

  removeGroup: (groupId) =>
    set((state) => ({
      groups: state.groups.filter((g) => g.id !== groupId),
    })),

  addFeed: (feed) => set((state) => ({ feeds: [...state.feeds, feed] })),

  updateFeed: (feedId, updates) =>
    set((state) => ({
      feeds: state.feeds.map((f) =>
        f.id === feedId ? { ...f, ...updates } : f
      ),
    })),

  removeFeed: (feedId) =>
    set((state) => ({
      feeds: state.feeds.filter((f) => f.id !== feedId),
    })),

  updateFeedGroup: (feedId, groupId) =>
    set((state) => ({
      feeds: state.feeds.map((f) =>
        f.id === feedId ? { ...f, group_id: groupId } : f
      ),
    })),

  moveFeedsToGroup: (fromGroupId, toGroupId) =>
    set((state) => ({
      feeds: state.feeds.map((f) =>
        f.group_id === fromGroupId ? { ...f, group_id: toGroupId } : f
      ),
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
