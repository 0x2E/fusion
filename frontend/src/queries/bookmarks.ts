import { useCallback, useMemo } from "react";
import {
  infiniteQueryOptions,
  queryOptions,
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  bookmarkAPI,
  type Bookmark,
  type Item,
  type ListAPIResponse,
  type SavedBookmarkRef,
} from "@/lib/api";
import { countFetchedRows } from "./article-list-logic";
import {
  buildSavedBookmarkLookup,
  resolveBookmarkFeedId,
  resolveSavedBookmarkId,
  removeSavedBookmarkRef,
  upsertSavedBookmarkRef,
} from "./bookmark-lookup-logic";
import { useFeedLookup } from "./feeds";
import { queryKeys } from "./keys";

type CachedItemPages = {
  pages: Array<{ data: Item[] }>;
};

function getCachedItemById(
  queryClient: ReturnType<typeof useQueryClient>,
  itemId: number,
): Item | undefined {
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

const bookmarkPageSize = 100;

export const bookmarkQueries = {
  refs: () =>
    queryOptions({
      queryKey: queryKeys.bookmarks.refs(),
      queryFn: async () => {
        const res = await bookmarkAPI.listSavedItemRefs();
        return res.data ?? [];
      },
      staleTime: Number.POSITIVE_INFINITY,
    }),
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

export function useSavedBookmarkRefs() {
  return useQuery(bookmarkQueries.refs());
}

export function useBookmarkLookup() {
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading: isBookmarksLoading,
  } = useBookmarks();
  const {
    data: savedRefs = [],
    isLoading: isSavedRefsLoading,
  } = useSavedBookmarkRefs();
  const pages = data?.pages ?? [];
  const bookmarks = useMemo(() => pages.flatMap((page) => page.data), [pages]);
  const fetchedCount = useMemo(() => countFetchedRows(pages), [pages]);
  const total = pages.at(-1)?.total ?? savedRefs.length;
  const savedLookup = useMemo(
    () => buildSavedBookmarkLookup(savedRefs),
    [savedRefs],
  );

  const byArticleId = useMemo(
    () => new Map(bookmarks.map((bookmark) => [resolveBookmarkItemId(bookmark), bookmark])),
    [bookmarks],
  );

  const isItemStarred = useCallback(
    (itemId: number) => resolveSavedBookmarkId({ savedLookup, loadedBookmarks: byArticleId, itemId }) !== undefined,
    [byArticleId, savedLookup],
  );

  const getBookmarkByItemId = useCallback(
    (itemId: number) => byArticleId.get(itemId),
    [byArticleId],
  );

  const getSavedBookmarkIdByItemId = useCallback(
    (itemId: number) =>
      resolveSavedBookmarkId({ savedLookup, loadedBookmarks: byArticleId, itemId }),
    [byArticleId, savedLookup],
  );

  return {
    bookmarks,
    fetchedCount,
    total,
    hasNextPage,
    fetchNextPage,
    isFetchingNextPage,
    isLoading: isBookmarksLoading || isSavedRefsLoading,
    isItemStarred,
    getBookmarkByItemId,
    getSavedBookmarkIdByItemId,
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
        (bookmark) => {
          const feedId = resolveBookmarkFeedId({
            bookmark,
            uniqueFeedIdByName,
            getCachedFeedId: (itemId) => getCachedItemById(queryClient, itemId)?.feed_id,
          });
          return feedId === filters.feedId;
        },
      );
    } else if (filters.groupId) {
      filtered = filtered.filter((bookmark) => {
        const feedId = resolveBookmarkFeedId({
          bookmark,
          uniqueFeedIdByName,
          getCachedFeedId: (itemId) => getCachedItemById(queryClient, itemId)?.feed_id,
        });
        return feedId > 0 && groupFeedIds?.has(feedId) === true;
      });
    }

    return filtered.map(
      (bookmark): Item => ({
        id: bookmark.item_id ?? -bookmark.id,
        feed_id: resolveBookmarkFeedId({
          bookmark,
          uniqueFeedIdByName,
          getCachedFeedId: (itemId) => getCachedItemById(queryClient, itemId)?.feed_id,
        }),
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
      qc.setQueryData(queryKeys.bookmarks.refs(), (old: SavedBookmarkRef[] | undefined) =>
        upsertSavedBookmarkRef(old, bookmark),
      );
      void qc.invalidateQueries({ queryKey: queryKeys.bookmarks.list() });
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
      qc.setQueryData(queryKeys.bookmarks.refs(), (old: SavedBookmarkRef[] | undefined) =>
        removeSavedBookmarkRef(old, bookmarkId),
      );
      void qc.invalidateQueries({ queryKey: queryKeys.bookmarks.list() });
    },
  });
}
