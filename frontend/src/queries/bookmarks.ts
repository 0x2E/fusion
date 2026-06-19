import { useCallback, useMemo } from "react";
import {
  useInfiniteQuery,
  useMutation,
  useQueryClient,
  type InfiniteData,
} from "@tanstack/react-query";
import {
  bookmarkAPI,
  type Bookmark,
  type Item,
  type ListAPIResponse,
} from "@/lib/api";
import {
  normalizeBookmarkFilters,
  queryKeys,
  type BookmarkFilters,
  type NormalizedBookmarkFilters,
} from "./keys";
import { useFeedLookup } from "./feeds";
import { usePreferencesStore } from "@/store";

// The lookup query (star icons + sidebar count) fetches a large first page so
// most users' bookmarks are covered in a single request. 100 is also the
// backend per-request cap, so it is the natural chunk size.
const BOOKMARK_LOOKUP_PAGE_SIZE = 100;

type BookmarkListResponse = ListAPIResponse<Bookmark>;
export type BookmarksInfiniteData = InfiniteData<BookmarkListResponse, number>;

export function resolveBookmarkItemId(bookmark: Bookmark): number {
  return bookmark.item_id ?? -bookmark.id;
}

function buildListBookmarksParams(
  filters: NormalizedBookmarkFilters,
  offset: number,
  pageSize: number,
) {
  const params: Parameters<typeof bookmarkAPI.list>[0] = {
    limit: pageSize,
    offset,
  };
  if (filters.feedId) params.feed_id = filters.feedId;
  if (filters.groupId) params.group_id = filters.groupId;
  return params;
}

// useBookmarks is the shared infinite query over bookmarks. Callers pass the
// filters and page size that fit their use case (lookup vs. starred list).
// Options are inlined (rather than spread from a helper) so React Query infers
// the pageParam type as `number`.
function useBookmarks(
  filters: BookmarkFilters,
  pageSize: number,
  enabled = true,
) {
  const normalized = normalizeBookmarkFilters(filters);
  return useInfiniteQuery({
    queryKey: [...queryKeys.bookmarks.lists(), normalized, pageSize],
    queryFn: async ({ pageParam }) =>
      bookmarkAPI.list(
        buildListBookmarksParams(normalized, pageParam, pageSize),
      ),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) => {
      const fetched = allPages.reduce((n, p) => n + p.data.length, 0);
      return fetched < lastPage.total ? fetched : undefined;
    },
    staleTime: Number.POSITIVE_INFINITY,
    enabled,
  });
}

// useBookmarkLookup powers star indicators and the sidebar starred count.
// It is intentionally unfiltered so star state is consistent across views;
// the first page (up to BOOKMARK_LOOKUP_PAGE_SIZE) covers the common case and
// `total` always reflects the true global count.
export function useBookmarkLookup() {
  const query = useBookmarks({}, BOOKMARK_LOOKUP_PAGE_SIZE);

  const bookmarks = useMemo(
    () => query.data?.pages.flatMap((p) => p.data) ?? [],
    [query.data],
  );
  const total = query.data?.pages.at(-1)?.total ?? 0;

  const byItemId = useMemo(
    () => new Map(bookmarks.map((b) => [resolveBookmarkItemId(b), b])),
    [bookmarks],
  );

  const isItemStarred = useCallback(
    (itemId: number) => byItemId.has(itemId),
    [byItemId],
  );

  const getBookmarkByItemId = useCallback(
    (itemId: number) => byItemId.get(itemId),
    [byItemId],
  );

  return { bookmarks, total, isItemStarred, getBookmarkByItemId };
}

export interface StarredItemsResult {
  items: Item[];
  bookmarks: Bookmark[];
  hasNextPage: boolean;
  isLoading: boolean;
  isFetchingNextPage: boolean;
  fetchNextPage: () => void;
}

// useStarredItems is the paginated, server-filtered starred list. Filtering by
// feed/group is pushed to the backend, so pagination is correct in every scope.
export function useStarredItems(
  filters: BookmarkFilters,
  enabled = true,
): StarredItemsResult {
  const pageSize = usePreferencesStore((state) => state.articlePageSize);
  const query = useBookmarks(filters, pageSize, enabled);

  const bookmarks = useMemo(
    () => query.data?.pages.flatMap((p) => p.data) ?? [],
    [query.data],
  );

  const items = useMemo<Item[]>(
    () =>
      bookmarks.map((bookmark) => ({
        id: bookmark.item_id ?? -bookmark.id,
        feed_id: bookmark.feed_id ?? 0,
        guid: bookmark.link || `bookmark:${bookmark.id}`,
        title: bookmark.title,
        link: bookmark.link,
        content: bookmark.content,
        pub_date: bookmark.pub_date,
        unread: bookmark.unread,
        created_at: bookmark.created_at,
      })),
    [bookmarks],
  );

  return {
    items,
    bookmarks,
    hasNextPage: query.hasNextPage,
    isLoading: query.isLoading,
    isFetchingNextPage: query.isFetchingNextPage,
    fetchNextPage: () => void query.fetchNextPage(),
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
      const itemId = resolveBookmarkItemId(bookmark);
      qc.setQueriesData<BookmarksInfiniteData>(
        { queryKey: queryKeys.bookmarks.lists() },
        (old) => {
          if (!old) return old;
          const pages = [...old.pages];
          const first = pages[0];
          if (!first) return old;

          const index = first.data.findIndex(
            (b) => resolveBookmarkItemId(b) === itemId,
          );
          if (index !== -1) {
            const newData = [...first.data];
            newData[index] = bookmark;
            pages[0] = { ...first, data: newData };
          } else {
            pages[0] = {
              ...first,
              data: [bookmark, ...first.data],
              total: first.total + 1,
            };
          }
          return { ...old, pages };
        },
      );
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: queryKeys.bookmarks.all });
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
      qc.setQueriesData<BookmarksInfiniteData>(
        { queryKey: queryKeys.bookmarks.lists() },
        (old) => {
          if (!old) return old;
          const pages = old.pages.map((page) => {
            const newData = page.data.filter((b) => b.id !== bookmarkId);
            if (newData.length === page.data.length) return page;
            return {
              ...page,
              data: newData,
              total: Math.max(0, page.total - 1),
            };
          });
          return { ...old, pages };
        },
      );
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: queryKeys.bookmarks.all });
    },
  });
}
