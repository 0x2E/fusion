import { useCallback, useMemo } from "react";
import {
  type QueryClient,
  type InfiniteData,
  infiniteQueryOptions,
  useInfiniteQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import {
  bookmarkAPI,
  type Bookmark,
  type Item,
  type ListAPIResponse,
} from "@/lib/api";
import { useFeedLookup } from "./feeds";
import { queryKeys } from "./keys";

type CachedItemPages = {
  pages: Array<{ data: Item[] }>;
};

function getCachedItemById(queryClient: QueryClient, itemId: number): Item | undefined {
  const detailItem = queryClient.getQueryData<Item>(queryKeys.items.detail(itemId));
  if (detailItem) {
    return detailItem;
  }

  const itemLists = queryClient.getQueriesData<CachedItemPages>({
    queryKey: queryKeys.items.lists(),
  });

  for (const [, data] of itemLists) {
    const item = data?.pages.flatMap((page) => page.data).find((entry) => entry.id === itemId);
    if (item) {
      return item;
    }
  }

  return undefined;
}

function resolveBookmarkItemId(bookmark: Bookmark): number {
  return bookmark.item_id ?? -bookmark.id;
}

export type BookmarkPage = ListAPIResponse<Bookmark> & { fetchedCount: number };
type BookmarkInfiniteData = InfiniteData<BookmarkPage, number>;

const bookmarkPageSize = 100;

function updateBookmarkPages(
  old: BookmarkInfiniteData | undefined,
  updater: (page: BookmarkPage, pageIndex: number) => BookmarkPage,
  totalDelta: number,
): BookmarkInfiniteData | undefined {
  if (!old) {
    return old;
  }

  return {
    ...old,
    pages: old.pages.map((page, index) => ({
      ...updater(page, index),
      total: Math.max(0, page.total + totalDelta),
      fetchedCount: page.fetchedCount,
    })),
  };
}

function upsertBookmarkInPages(
  old: BookmarkInfiniteData | undefined,
  bookmark: Bookmark,
): BookmarkInfiniteData | undefined {
  if (!old) {
    return old;
  }

  const itemId = resolveBookmarkItemId(bookmark);
  const exists = old.pages.some((page) =>
    page.data.some(
      (entry) => resolveBookmarkItemId(entry) === itemId || entry.id === bookmark.id,
    ),
  );

  return updateBookmarkPages(
    old,
    (page, index) => {
      const existingIndex = page.data.findIndex(
        (entry) => resolveBookmarkItemId(entry) === itemId || entry.id === bookmark.id,
      );
      if (existingIndex !== -1) {
        const data = [...page.data];
        data[existingIndex] = bookmark;
        return { ...page, data };
      }

      if (!exists && index === 0) {
        return { ...page, data: [bookmark, ...page.data] };
      }

      return page;
    },
    exists ? 0 : 1,
  );
}

function removeBookmarkFromPages(
  old: BookmarkInfiniteData | undefined,
  bookmarkId: number,
): BookmarkInfiniteData | undefined {
  if (!old) {
    return old;
  }

  const removed = old.pages.some((page) =>
    page.data.some((entry) => entry.id === bookmarkId),
  );
  if (!removed) {
    return old;
  }

  return updateBookmarkPages(
    old,
    (page) => ({
      ...page,
      data: page.data.filter((entry) => entry.id !== bookmarkId),
    }),
    -1,
  );
}

function resolveBookmarkFeedId(
  bookmark: Bookmark,
  queryClient: QueryClient,
  uniqueFeedIdByName: Map<string, number>,
): number {
  if (bookmark.item_id && bookmark.item_id > 0) {
    const item = getCachedItemById(queryClient, bookmark.item_id);
    if (item) {
      return item.feed_id;
    }
  }

  return uniqueFeedIdByName.get(bookmark.feed_name) ?? 0;
}

export const bookmarkQueries = {
  list: () =>
    infiniteQueryOptions({
      queryKey: queryKeys.bookmarks.list(),
      queryFn: async ({ pageParam }) => {
        const page = await bookmarkAPI.list(bookmarkPageSize, pageParam);
        return {
          ...page,
          fetchedCount: page.data.length,
        };
      },
      initialPageParam: 0,
      getNextPageParam: (lastPage, allPages) => {
        const fetched = allPages.reduce((count, page) => count + page.fetchedCount, 0);
        return fetched < lastPage.total ? fetched : undefined;
      },
      staleTime: Number.POSITIVE_INFINITY,
    }),
};

export async function fetchBookmarkPage(offset: number): Promise<BookmarkPage> {
  const page = await bookmarkAPI.list(bookmarkPageSize, offset);
  return {
    ...page,
    fetchedCount: page.data.length,
  };
}

export function useBookmarks() {
  return useInfiniteQuery(bookmarkQueries.list());
}

export function useBookmarkLookup() {
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
  } = useBookmarks();
  const pages = data?.pages ?? [];
  const bookmarks = useMemo(() => pages.flatMap((page) => page.data), [pages]);
  const total = pages.at(-1)?.total ?? 0;

  const byArticleId = useMemo(
    () => new Map(bookmarks.map((bookmark) => [resolveBookmarkItemId(bookmark), bookmark])),
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

  return {
    bookmarks,
    total,
    hasNextPage,
    fetchNextPage,
    isFetchingNextPage,
    isLoading,
    isItemStarred,
    getBookmarkByItemId,
  };
}

export function useStarredItems(filters: {
  feedId: number | null;
  groupId: number | null;
}) {
  const queryClient = useQueryClient();
  const bookmarksQuery = useBookmarkLookup();
  const { feeds, getFeedsByGroup } = useFeedLookup();

  const articles = useMemo(() => {
    const feedNameCounts = new Map<string, number>();
    for (const feed of feeds) {
      feedNameCounts.set(feed.name, (feedNameCounts.get(feed.name) ?? 0) + 1);
    }

    const uniqueFeedIdByName = new Map(
      feeds
        .filter((feed) => (feedNameCounts.get(feed.name) ?? 0) === 1)
        .map((feed) => [feed.name, feed.id] as const),
    );
    const groupFeedIds =
      filters.groupId === null
        ? null
        : new Set(getFeedsByGroup(filters.groupId).map((feed) => feed.id));

    let filtered = bookmarksQuery.bookmarks;

    if (filters.feedId) {
      filtered = filtered.filter(
        (bookmark) =>
          resolveBookmarkFeedId(bookmark, queryClient, uniqueFeedIdByName) ===
          filters.feedId,
      );
    } else if (filters.groupId) {
      filtered = filtered.filter((bookmark) => {
        const feedId = resolveBookmarkFeedId(bookmark, queryClient, uniqueFeedIdByName);
        return feedId > 0 && groupFeedIds?.has(feedId) === true;
      });
    }

    return filtered.map(
      (bookmark): Item => ({
        id: bookmark.item_id ?? -bookmark.id,
        feed_id: resolveBookmarkFeedId(bookmark, queryClient, uniqueFeedIdByName),
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
    bookmarksQuery.bookmarks,
    feeds,
    filters.feedId,
    filters.groupId,
    getFeedsByGroup,
    queryClient,
  ]);

  return {
    articles,
    hasNextPage: bookmarksQuery.hasNextPage,
    fetchNextPage: bookmarksQuery.fetchNextPage,
    isFetchingNextPage: bookmarksQuery.isFetchingNextPage,
    isLoading: bookmarksQuery.isLoading,
  };
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
      qc.setQueryData(queryKeys.bookmarks.list(), (old: BookmarkInfiniteData | undefined) =>
        upsertBookmarkInPages(old, bookmark),
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
      qc.setQueryData(queryKeys.bookmarks.list(), (old: BookmarkInfiniteData | undefined) =>
        removeBookmarkFromPages(old, bookmarkId),
      );
    },
  });
}
