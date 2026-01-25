import { useEffect, useCallback, useMemo } from "react";
import { itemAPI, type ListItemsParams } from "@/lib/api";
import { useDataStore, useUIStore } from "@/store";

export function useArticles() {
  const {
    items,
    isLoadingItems,
    itemsError,
    setItems,
    setLoadingItems,
    setItemsError,
    markItemRead,
    markItemUnread,
    markItemsRead,
    getFeedById,
    isItemStarred,
  } = useDataStore();

  const { selectedFeedId, selectedGroupId, articleFilter } = useUIStore();

  const fetchArticles = useCallback(
    async (params?: ListItemsParams) => {
      setLoadingItems(true);
      setItemsError(null);
      try {
        const response = await itemAPI.list({
          limit: 100,
          order_by: "pub_date:desc",
          ...params,
        });
        setItems(response.data);
      } catch (error) {
        setItemsError(error instanceof Error ? error.message : "Failed to load articles");
      } finally {
        setLoadingItems(false);
      }
    },
    [setItems, setLoadingItems, setItemsError]
  );

  const refresh = useCallback(async () => {
    const params: ListItemsParams = {};
    if (selectedFeedId) params.feed_id = selectedFeedId;
    if (selectedGroupId) params.group_id = selectedGroupId;
    if (articleFilter === "unread") params.unread = true;
    await fetchArticles(params);
  }, [fetchArticles, selectedFeedId, selectedGroupId, articleFilter]);

  useEffect(() => {
    const params: ListItemsParams = {};
    if (selectedFeedId) params.feed_id = selectedFeedId;
    if (selectedGroupId) params.group_id = selectedGroupId;
    if (articleFilter === "unread") params.unread = true;
    fetchArticles(params);
  }, [selectedFeedId, selectedGroupId, articleFilter, fetchArticles]);

  const filteredArticles = useMemo(() => {
    let result = items;

    if (articleFilter === "starred") {
      result = result.filter((item) => isItemStarred(item.id));
    }

    return result;
  }, [items, articleFilter, isItemStarred]);

  const markAsRead = useCallback(
    async (itemId: number) => {
      try {
        await itemAPI.markRead({ ids: [itemId] });
        markItemRead(itemId);
      } catch (error) {
        console.error("Failed to mark as read:", error);
      }
    },
    [markItemRead]
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
    [markItemUnread]
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
      const item = items.find((i) => i.id === itemId);
      if (!item) return null;

      const feed = getFeedById(item.feed_id);
      return {
        ...item,
        feedName: feed?.name ?? "Unknown",
        isStarred: isItemStarred(item.id),
      };
    },
    [items, getFeedById, isItemStarred]
  );

  return {
    articles: filteredArticles,
    isLoading: isLoadingItems,
    error: itemsError,
    refresh,
    markAsRead,
    markAsUnread,
    markAllAsRead,
    getArticleWithMeta,
  };
}
