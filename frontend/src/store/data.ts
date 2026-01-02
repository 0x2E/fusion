import { create } from "zustand";
import { groupAPI, feedAPI, itemAPI, bookmarkAPI } from "../lib/api";
import type {
  Group,
  Feed,
  Item,
  Bookmark,
  CreateGroupRequest,
  UpdateGroupRequest,
  CreateFeedRequest,
  UpdateFeedRequest,
  ListItemsParams,
  CreateBookmarkRequest,
} from "../lib/api";
import { toast } from "sonner";

interface DataState {
  // Data
  groups: Group[];
  feeds: Feed[];
  items: Item[];
  bookmarks: Bookmark[];
  itemsTotal: number;
  bookmarksTotal: number;

  // Loading states
  groupsLoading: boolean;
  feedsLoading: boolean;
  itemsLoading: boolean;
  bookmarksLoading: boolean;

  // Fetch methods
  fetchGroups: () => Promise<void>;
  fetchFeeds: () => Promise<void>;
  fetchItems: (params?: ListItemsParams) => Promise<void>;
  fetchBookmarks: (limit?: number, offset?: number) => Promise<void>;

  // Group CRUD
  createGroup: (data: CreateGroupRequest) => Promise<Group>;
  updateGroup: (id: number, data: UpdateGroupRequest) => Promise<void>;
  deleteGroup: (id: number) => Promise<void>;

  // Feed CRUD
  createFeed: (data: CreateFeedRequest) => Promise<Feed>;
  updateFeed: (id: number, data: UpdateFeedRequest) => Promise<void>;
  deleteFeed: (id: number) => Promise<void>;
  refreshFeeds: () => Promise<void>;

  // Item operations
  markItemsRead: (ids: number[]) => Promise<void>;
  markItemsUnread: (ids: number[]) => Promise<void>;
  deleteItem: (id: number) => Promise<void>;

  // Bookmark operations
  createBookmark: (data: CreateBookmarkRequest) => Promise<Bookmark>;
  deleteBookmark: (id: number) => Promise<void>;

  // Reset all data
  reset: () => void;
}

export const useDataStore = create<DataState>((set, get) => ({
  // Initial state
  groups: [],
  feeds: [],
  items: [],
  bookmarks: [],
  itemsTotal: 0,
  bookmarksTotal: 0,
  groupsLoading: false,
  feedsLoading: false,
  itemsLoading: false,
  bookmarksLoading: false,

  // Fetch methods
  fetchGroups: async () => {
    set({ groupsLoading: true });
    try {
      const response = await groupAPI.list();
      set({ groups: response.data, groupsLoading: false });
    } catch (error) {
      set({ groupsLoading: false });
      toast.error(
        error instanceof Error ? error.message : "Failed to fetch groups"
      );
      throw error;
    }
  },

  fetchFeeds: async () => {
    set({ feedsLoading: true });
    try {
      const response = await feedAPI.list();
      set({ feeds: response.data, feedsLoading: false });
    } catch (error) {
      set({ feedsLoading: false });
      toast.error(
        error instanceof Error ? error.message : "Failed to fetch feeds"
      );
      throw error;
    }
  },

  fetchItems: async (params?: ListItemsParams) => {
    set({ itemsLoading: true });
    try {
      const response = await itemAPI.list(params);
      set({
        items: response.data,
        itemsTotal: response.total,
        itemsLoading: false,
      });
    } catch (error) {
      set({ itemsLoading: false });
      toast.error(
        error instanceof Error ? error.message : "Failed to fetch items"
      );
      throw error;
    }
  },

  fetchBookmarks: async (limit = 50, offset = 0) => {
    set({ bookmarksLoading: true });
    try {
      const response = await bookmarkAPI.list(limit, offset);
      set({
        bookmarks: response.data,
        bookmarksTotal: response.total,
        bookmarksLoading: false,
      });
    } catch (error) {
      set({ bookmarksLoading: false });
      toast.error(
        error instanceof Error ? error.message : "Failed to fetch bookmarks"
      );
      throw error;
    }
  },

  // Group CRUD
  createGroup: async (data: CreateGroupRequest) => {
    try {
      const response = await groupAPI.create(data);
      if (!response.data) {
        throw new Error("No data returned from create group");
      }
      // Optimistic update
      set((state) => ({
        groups: [...state.groups, response.data!],
      }));
      toast.success("Group created successfully");
      return response.data;
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Failed to create group"
      );
      throw error;
    }
  },

  updateGroup: async (id: number, data: UpdateGroupRequest) => {
    const { groups } = get();
    const originalGroup = groups.find((g) => g.id === id);

    // Optimistic update
    set({
      groups: groups.map((g) => (g.id === id ? { ...g, ...data } : g)),
    });

    try {
      await groupAPI.update(id, data);
      toast.success("Group updated successfully");
    } catch (error) {
      // Rollback on error
      if (originalGroup) {
        set({
          groups: groups.map((g) => (g.id === id ? originalGroup : g)),
        });
      }
      toast.error(
        error instanceof Error ? error.message : "Failed to update group"
      );
      throw error;
    }
  },

  deleteGroup: async (id: number) => {
    const { groups } = get();
    const originalGroups = [...groups];

    // Optimistic update
    set({ groups: groups.filter((g) => g.id !== id) });

    try {
      await groupAPI.delete(id);
      toast.success("Group deleted successfully");
    } catch (error) {
      // Rollback on error
      set({ groups: originalGroups });
      toast.error(
        error instanceof Error ? error.message : "Failed to delete group"
      );
      throw error;
    }
  },

  // Feed CRUD
  createFeed: async (data: CreateFeedRequest) => {
    try {
      const response = await feedAPI.create(data);
      if (!response.data) {
        throw new Error("No data returned from create feed");
      }
      // Optimistic update
      set((state) => ({
        feeds: [...state.feeds, response.data!],
      }));
      toast.success("Feed created successfully");
      return response.data;
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Failed to create feed"
      );
      throw error;
    }
  },

  updateFeed: async (id: number, data: UpdateFeedRequest) => {
    const { feeds } = get();
    const originalFeed = feeds.find((f) => f.id === id);

    // Optimistic update
    set({
      feeds: feeds.map((f) => (f.id === id ? { ...f, ...data } : f)),
    });

    try {
      await feedAPI.update(id, data);
      toast.success("Feed updated successfully");
    } catch (error) {
      // Rollback on error
      if (originalFeed) {
        set({
          feeds: feeds.map((f) => (f.id === id ? originalFeed : f)),
        });
      }
      toast.error(
        error instanceof Error ? error.message : "Failed to update feed"
      );
      throw error;
    }
  },

  deleteFeed: async (id: number) => {
    const { feeds } = get();
    const originalFeeds = [...feeds];

    // Optimistic update
    set({ feeds: feeds.filter((f) => f.id !== id) });

    try {
      await feedAPI.delete(id);
      toast.success("Feed deleted successfully");
    } catch (error) {
      // Rollback on error
      set({ feeds: originalFeeds });
      toast.error(
        error instanceof Error ? error.message : "Failed to delete feed"
      );
      throw error;
    }
  },

  refreshFeeds: async () => {
    try {
      await feedAPI.refresh();
      toast.success("Feeds refresh triggered");
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Failed to refresh feeds"
      );
      throw error;
    }
  },

  // Item operations
  markItemsRead: async (ids: number[]) => {
    const { items } = get();
    const originalItems = [...items];

    // Optimistic update
    set({
      items: items.map((item) =>
        ids.includes(item.id) ? { ...item, unread: false } : item
      ),
    });

    try {
      await itemAPI.markRead({ ids });
      toast.success(`Marked ${ids.length} item(s) as read`);
    } catch (error) {
      // Rollback on error
      set({ items: originalItems });
      toast.error(
        error instanceof Error ? error.message : "Failed to mark items as read"
      );
      throw error;
    }
  },

  markItemsUnread: async (ids: number[]) => {
    const { items } = get();
    const originalItems = [...items];

    // Optimistic update
    set({
      items: items.map((item) =>
        ids.includes(item.id) ? { ...item, unread: true } : item
      ),
    });

    try {
      await itemAPI.markUnread({ ids });
      toast.success(`Marked ${ids.length} item(s) as unread`);
    } catch (error) {
      // Rollback on error
      set({ items: originalItems });
      toast.error(
        error instanceof Error
          ? error.message
          : "Failed to mark items as unread"
      );
      throw error;
    }
  },

  deleteItem: async (id: number) => {
    const { items } = get();
    const originalItems = [...items];

    // Optimistic update
    set({ items: items.filter((item) => item.id !== id) });

    try {
      await itemAPI.delete(id);
      toast.success("Item deleted successfully");
    } catch (error) {
      // Rollback on error
      set({ items: originalItems });
      toast.error(
        error instanceof Error ? error.message : "Failed to delete item"
      );
      throw error;
    }
  },

  // Bookmark operations
  createBookmark: async (data: CreateBookmarkRequest) => {
    try {
      const response = await bookmarkAPI.create(data);
      if (!response.data) {
        throw new Error("No data returned from create bookmark");
      }
      // Optimistic update
      set((state) => ({
        bookmarks: [response.data!, ...state.bookmarks],
      }));
      toast.success("Bookmark created successfully");
      return response.data;
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Failed to create bookmark"
      );
      throw error;
    }
  },

  deleteBookmark: async (id: number) => {
    const { bookmarks } = get();
    const originalBookmarks = [...bookmarks];

    // Optimistic update
    set({ bookmarks: bookmarks.filter((b) => b.id !== id) });

    try {
      await bookmarkAPI.delete(id);
      toast.success("Bookmark deleted successfully");
    } catch (error) {
      // Rollback on error
      set({ bookmarks: originalBookmarks });
      toast.error(
        error instanceof Error ? error.message : "Failed to delete bookmark"
      );
      throw error;
    }
  },

  reset: () => {
    set({
      groups: [],
      feeds: [],
      items: [],
      bookmarks: [],
      itemsTotal: 0,
      bookmarksTotal: 0,
    });
  },
}));
