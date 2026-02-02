import { useEffect, useCallback, useMemo, useState } from "react";
import { itemAPI, type ListItemsParams, type Item } from "@/lib/api";
import { useDataStore } from "@/store";
import { useUrlState } from "./use-url-state";

const PAGE_SIZE = 10;

export function useArticles() {
  const {
    items,
    itemsTotal,
    isLoadingItems,
    itemsError,
    setItems,
    appendItems,
    setItemsTotal,
    setLoadingItems,
    setItemsError,
    markItemRead,
    markItemUnread,
    markItemsRead,
    getFeedById,
    isItemStarred,
    bookmarks,
    feeds,
  } = useDataStore();

  const { selectedFeedId, selectedGroupId, articleFilter } = useUrlState();
  const [isLoadingMore, setIsLoadingMore] = useState(false);

  const fetchArticles = useCallback(
    async (params?: ListItemsParams) => {
      setLoadingItems(true);
      setItemsError(null);
      try {
        const response = await itemAPI.list({
          limit: PAGE_SIZE,
          order_by: "pub_date:desc",
          ...params,
        });
        setItems(response.data);
        setItemsTotal(response.total);
      } catch (error) {
        setItemsError(
          error instanceof Error ? error.message : "Failed to load articles",
        );
      } finally {
        setLoadingItems(false);
      }
    },
    [setItems, setItemsTotal, setLoadingItems, setItemsError],
  );

  const refresh = useCallback(async () => {
    const params: ListItemsParams = {};
    if (selectedFeedId) params.feed_id = selectedFeedId;
    if (selectedGroupId) params.group_id = selectedGroupId;
    if (articleFilter === "unread") params.unread = true;
    await fetchArticles(params);
  }, [fetchArticles, selectedFeedId, selectedGroupId, articleFilter]);

  const loadMore = useCallback(async () => {
    if (isLoadingMore || items.length >= itemsTotal) return;

    setIsLoadingMore(true);
    try {
      const params: ListItemsParams = {
        limit: PAGE_SIZE,
        offset: items.length,
        order_by: "pub_date:desc",
      };
      if (selectedFeedId) params.feed_id = selectedFeedId;
      if (selectedGroupId) params.group_id = selectedGroupId;
      if (articleFilter === "unread") params.unread = true;

      const response = await itemAPI.list(params);
      appendItems(response.data);
      setItemsTotal(response.total);
    } catch (error) {
      console.error("Failed to load more articles:", error);
    } finally {
      setIsLoadingMore(false);
    }
  }, [
    isLoadingMore,
    items.length,
    itemsTotal,
    selectedFeedId,
    selectedGroupId,
    articleFilter,
    appendItems,
    setItemsTotal,
  ]);

  // In starred mode, all bookmarks are loaded at once, so no pagination needed
  const hasMore =
    articleFilter === "starred" ? false : items.length < itemsTotal;

  useEffect(() => {
    const params: ListItemsParams = {};
    if (selectedFeedId) params.feed_id = selectedFeedId;
    if (selectedGroupId) params.group_id = selectedGroupId;
    if (articleFilter === "unread") params.unread = true;
    fetchArticles(params);
  }, [selectedFeedId, selectedGroupId, articleFilter, fetchArticles]);

  const filteredArticles = useMemo(() => {
    if (articleFilter === "starred") {
      // For starred filter, use bookmarks data directly
      // This ensures all starred items are shown, not just those in current items page
      let filteredBookmarks = bookmarks;

      // If a specific feed is selected, filter by feed
      if (selectedFeedId) {
        const feed = getFeedById(selectedFeedId);
        if (feed) {
          filteredBookmarks = bookmarks.filter(
            (b) => b.feed_name === feed.name,
          );
        }
      }

      // If a group is selected, filter by feeds in that group
      if (selectedGroupId) {
        const groupFeeds = feeds.filter((f) => f.group_id === selectedGroupId);
        const feedNames = new Set(groupFeeds.map((f) => f.name));
        filteredBookmarks = bookmarks.filter((b) => feedNames.has(b.feed_name));
      }

      // Convert bookmarks to Item-like objects
      return filteredBookmarks.map(
        (b): Item => ({
          id: b.item_id ?? -b.id, // Use negative bookmark id if item_id is null
          feed_id: feeds.find((f) => f.name === b.feed_name)?.id ?? 0,
          guid: b.link,
          title: b.title,
          link: b.link,
          content: b.content,
          pub_date: b.pub_date,
          unread: false, // Starred items displayed as read
          created_at: b.created_at,
        }),
      );
    }

    return items;
  }, [
    items,
    articleFilter,
    bookmarks,
    selectedFeedId,
    selectedGroupId,
    getFeedById,
    feeds,
  ]);

  const markAsRead = useCallback(
    async (itemId: number) => {
      try {
        await itemAPI.markRead({ ids: [itemId] });
        markItemRead(itemId);
      } catch (error) {
        console.error("Failed to mark as read:", error);
      }
    },
    [markItemRead],
  );

  const markAsUnread = useCallback(
    async (itemId: number) => {
      try {
        await itemAPI.markUnread({ ids: [itemId] });
        markItemUnread(itemId);
      } catch (error) {
        console.error("Failed to mark as unread:", error);
      }
    },
    [markItemUnread],
  );

  const markAllAsRead = useCallback(async () => {
    const unreadIds = filteredArticles.filter((a) => a.unread).map((a) => a.id);
    if (unreadIds.length === 0) return;

    try {
      await itemAPI.markRead({ ids: unreadIds });
      markItemsRead(unreadIds);
    } catch (error) {
      console.error("Failed to mark all as read:", error);
    }
  }, [filteredArticles, markItemsRead]);

  const getArticleWithMeta = useCallback(
    (itemId: number) => {
      // First try to find in items
      let item = items.find((i) => i.id === itemId);

      // If not found and in starred mode, look in filteredArticles (from bookmarks)
      if (!item && articleFilter === "starred") {
        item = filteredArticles.find((i) => i.id === itemId);
      }

      if (!item) return null;

      const feed = getFeedById(item.feed_id);
      return {
        ...item,
        feedName: feed?.name ?? "Unknown",
        isStarred: isItemStarred(item.id),
      };
    },
    [items, filteredArticles, articleFilter, getFeedById, isItemStarred],
  );

  return {
    articles: filteredArticles,
    isLoading: isLoadingItems,
    isLoadingMore,
    hasMore,
    error: itemsError,
    refresh,
    loadMore,
    markAsRead,
    markAsUnread,
    markAllAsRead,
    getArticleWithMeta,
  };
}
