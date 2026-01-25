import { useEffect, useCallback, useRef } from "react";
import { bookmarkAPI, type Item } from "@/lib/api";
import { useDataStore } from "@/store";

export function useStarred() {
  const didInitRef = useRef(false);
  const {
    bookmarks,
    isLoadingBookmarks,
    bookmarksError,
    setBookmarks,
    setLoadingBookmarks,
    setBookmarksError,
    addBookmark,
    removeBookmark,
    isItemStarred,
    getBookmarkByItemId,
    getFeedById,
  } = useDataStore();

  const fetchBookmarks = useCallback(async () => {
    setLoadingBookmarks(true);
    setBookmarksError(null);
    try {
      const response = await bookmarkAPI.list(100, 0);
      setBookmarks(response.data);
    } catch (error) {
      setBookmarksError(error instanceof Error ? error.message : "Failed to load starred");
    } finally {
      setLoadingBookmarks(false);
    }
  }, [setBookmarks, setLoadingBookmarks, setBookmarksError]);

  useEffect(() => {
    // Prevent infinite re-fetch loops on failed requests (or valid empty responses).
    // Only perform the initial load attempt once per mount.
    if (didInitRef.current) return;
    didInitRef.current = true;

    if (bookmarks.length === 0) {
      fetchBookmarks();
    }
  }, [bookmarks.length, fetchBookmarks]);

  const starArticle = useCallback(
    async (item: Item) => {
      const feed = getFeedById(item.feed_id);
      try {
        const response = await bookmarkAPI.create({
          item_id: item.id,
          link: item.link,
          title: item.title,
          content: item.content,
          pub_date: item.pub_date,
          feed_name: feed?.name ?? "Unknown",
        });
        if (response.data) {
          addBookmark(response.data);
        }
      } catch (error) {
        console.error("Failed to star article:", error);
        throw error;
      }
    },
    [getFeedById, addBookmark]
  );

  const unstarArticle = useCallback(
    async (itemId: number) => {
      const bookmark = getBookmarkByItemId(itemId);
      if (!bookmark) return;

      try {
        await bookmarkAPI.delete(bookmark.id);
        removeBookmark(bookmark.id);
      } catch (error) {
        console.error("Failed to unstar article:", error);
        throw error;
      }
    },
    [getBookmarkByItemId, removeBookmark]
  );

  const toggleStar = useCallback(
    async (item: Item) => {
      if (isItemStarred(item.id)) {
        await unstarArticle(item.id);
      } else {
        await starArticle(item);
      }
    },
    [isItemStarred, starArticle, unstarArticle]
  );

  return {
    bookmarks,
    isLoading: isLoadingBookmarks,
    error: bookmarksError,
    refresh: fetchBookmarks,
    starArticle,
    unstarArticle,
    toggleStar,
    isStarred: isItemStarred,
  };
}
