import { useCallback, useMemo } from "react";
import {
  queryOptions,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { bookmarkAPI, type Bookmark, type Item } from "@/lib/api";
import { useArticleSessionStore } from "@/store/article-session";
import { queryKeys } from "./keys";
import { useFeedLookup } from "./feeds";

function resolveBookmarkItemId(bookmark: Bookmark): number {
  return bookmark.item_id ?? -bookmark.id;
}

export const bookmarkQueries = {
  list: () =>
    queryOptions({
      queryKey: queryKeys.bookmarks.list(),
      queryFn: async () => {
        const res = await bookmarkAPI.list(100, 0);
        return res.data;
      },
      staleTime: Number.POSITIVE_INFINITY,
    }),
};

export function useBookmarks() {
  return useQuery(bookmarkQueries.list());
}

export function useBookmarkLookup() {
  const { data: bookmarks = [] } = useBookmarks();
  const starredOverrides = useArticleSessionStore((s) => s.starredOverrides);

  const byArticleId = useMemo(
    () => new Map(bookmarks.map((b) => [resolveBookmarkItemId(b), b])),
    [bookmarks],
  );

  const isItemStarred = useCallback(
    (itemId: number) => starredOverrides[itemId] ?? byArticleId.has(itemId),
    [byArticleId, starredOverrides],
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
  const setStarredOverride = useArticleSessionStore((s) => s.setStarredOverride);

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
      const itemId = resolveBookmarkItemId(bookmark);
      qc.setQueryData(
        queryKeys.bookmarks.list(),
        (old: Bookmark[] | undefined) => {
          if (!old) return [bookmark];

          const index = old.findIndex((b) => resolveBookmarkItemId(b) === itemId);
          if (index === -1) {
            return [bookmark, ...old];
          }

          const next = [...old];
          next[index] = bookmark;
          return next;
        },
      );
      setStarredOverride(itemId, true);
    },
  });
}

export function useDeleteBookmark() {
  const qc = useQueryClient();
  const setStarredOverride = useArticleSessionStore((s) => s.setStarredOverride);

  return useMutation({
    mutationFn: async (bookmarkId: number) => {
      await bookmarkAPI.delete(bookmarkId);
      return bookmarkId;
    },
    onSuccess: (bookmarkId) => {
      const bookmark = qc
        .getQueryData<Bookmark[]>(queryKeys.bookmarks.list())
        ?.find((b) => b.id === bookmarkId);
      if (!bookmark) {
        return;
      }

      setStarredOverride(resolveBookmarkItemId(bookmark), false);
    },
  });
}
