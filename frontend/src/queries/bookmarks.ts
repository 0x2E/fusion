import { useCallback, useMemo } from "react";
import {
  queryOptions,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { bookmarkAPI, type Bookmark, type Item } from "@/lib/api";
import { queryKeys } from "./keys";
import { useFeedLookup } from "./feeds";

export const bookmarkQueries = {
  list: () =>
    queryOptions({
      queryKey: queryKeys.bookmarks.list(),
      queryFn: async () => {
        const res = await bookmarkAPI.list(100, 0);
        return res.data;
      },
    }),
};

export function useBookmarks() {
  return useQuery(bookmarkQueries.list());
}

export function useBookmarkLookup() {
  const { data: bookmarks = [] } = useBookmarks();

  const byArticleId = useMemo(
    () =>
      new Map(
        bookmarks.map((b) => [b.item_id ?? -b.id, b]),
      ),
    [bookmarks],
  );

  const isItemStarred = useCallback(
    (itemId: number) => byArticleId.has(itemId),
    [byArticleId],
  );

  const getBookmarkByItemId = useCallback(
    (itemId: number) => byArticleId.get(itemId),
    [byArticleId],
  );

  return { bookmarks, isItemStarred, getBookmarkByItemId };
}

export function useStarredItems(filters: {
  feedId: number | null;
  groupId: number | null;
}) {
  const { bookmarks } = useBookmarkLookup();
  const { feeds, getFeedById, getFeedsByGroup } = useFeedLookup();

  return useMemo(() => {
    let filtered = bookmarks;

    if (filters.feedId) {
      const feed = getFeedById(filters.feedId);
      if (!feed) {
        return [];
      }
      filtered = filtered.filter((b) => b.feed_name === feed.name);
    } else if (filters.groupId) {
      const feedNames = new Set(getFeedsByGroup(filters.groupId).map((f) => f.name));
      filtered = filtered.filter((b) => feedNames.has(b.feed_name));
    }

    const feedIdByName = new Map(feeds.map((f) => [f.name, f.id]));

    return filtered.map(
      (bookmark): Item => ({
        id: bookmark.item_id ?? -bookmark.id,
        feed_id: feedIdByName.get(bookmark.feed_name) ?? 0,
        guid: bookmark.link || `bookmark:${bookmark.id}`,
        title: bookmark.title,
        link: bookmark.link,
        content: bookmark.content,
        pub_date: bookmark.pub_date,
        unread: false,
        created_at: bookmark.created_at,
      }),
    );
  }, [
    bookmarks,
    filters.feedId,
    filters.groupId,
    feeds,
    getFeedById,
    getFeedsByGroup,
  ]);
}

export function useCreateBookmark() {
  const qc = useQueryClient();
  const { getFeedById } = useFeedLookup();

  return useMutation({
    mutationFn: async (item: Item) => {
      const feed = getFeedById(item.feed_id);
      const res = await bookmarkAPI.create({
        item_id: item.id,
        link: item.link,
        title: item.title,
        content: item.content,
        pub_date: item.pub_date,
        feed_name: feed?.name ?? "Unknown",
      });
      return res.data!;
    },
    onSuccess: (bookmark) => {
      qc.setQueryData(
        queryKeys.bookmarks.list(),
        (old: Bookmark[] | undefined) =>
          old ? [bookmark, ...old] : [bookmark],
      );
    },
  });
}

export function useDeleteBookmark() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (bookmarkId: number) => {
      await bookmarkAPI.delete(bookmarkId);
      return bookmarkId;
    },
    onSuccess: (bookmarkId) => {
      qc.setQueryData(
        queryKeys.bookmarks.list(),
        (old: Bookmark[] | undefined) =>
          old?.filter((b) => b.id !== bookmarkId),
      );
    },
  });
}
