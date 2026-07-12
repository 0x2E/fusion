import {
  infiniteQueryOptions,
  queryOptions,
  type QueryClient,
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
  type InfiniteData,
} from "@tanstack/react-query";
import { itemAPI, type Feed, type Item, type ListItemsParams } from "@/lib/api";
import {
  normalizeItemFilters,
  queryKeys,
  type ItemFilters,
  type NormalizedItemFilters,
} from "./keys";
import { usePreferencesStore } from "@/store";
import type { BookmarksInfiniteData } from "./bookmarks";

type ItemListResponse = Awaited<ReturnType<typeof itemAPI.list>>;
type ItemsInfiniteData = InfiniteData<ItemListResponse, number>;
type ItemsMutationContext = {
  prevItemLists: Array<[readonly unknown[], ItemsInfiniteData | undefined]>;
  prevItemDetails: Array<readonly [number, Item | undefined]>;
  prevFeeds: Feed[] | undefined;
  prevBookmarkLists: Array<
    [readonly unknown[], BookmarksInfiniteData | undefined]
  >;
};

function buildListItemsParams(
  filters: NormalizedItemFilters,
  offset: number,
  pageSize: number,
): ListItemsParams {
  const params: ListItemsParams = {
    limit: pageSize,
    offset,
    order_by: "pub_date:desc",
  };

  if (filters.feedId) params.feed_id = filters.feedId;
  if (filters.groupId) params.group_id = filters.groupId;
  if (filters.unread) params.unread = true;

  return params;
}

export const itemQueries = {
  list: (filters: ItemFilters, pageSize: number) => {
    const normalizedFilters = normalizeItemFilters(filters);

    return infiniteQueryOptions({
      queryKey: [...queryKeys.items.lists(), normalizedFilters, pageSize],
      queryFn: async ({ pageParam }) =>
        itemAPI.list(buildListItemsParams(normalizedFilters, pageParam, pageSize)),
      initialPageParam: 0,
      getNextPageParam: (lastPage, allPages) => {
        const fetched = allPages.reduce((n, p) => n + p.data.length, 0);
        return fetched < lastPage.total ? fetched : undefined;
      },
    });
  },
  detail: (itemId: number) =>
    queryOptions({
      queryKey: queryKeys.items.detail(itemId),
      queryFn: async () => {
        const res = await itemAPI.get(itemId);
        return res.data;
      },
    }),
};

export function useItems(filters: ItemFilters, enabled = true) {
  const articlePageSize = usePreferencesStore((state) => state.articlePageSize);
  return useInfiniteQuery({ ...itemQueries.list(filters, articlePageSize), enabled });
}

export function useItem(itemId: number | null, enabled = true) {
  return useQuery({
    ...itemQueries.detail(itemId ?? 0),
    enabled: enabled && itemId !== null && itemId > 0,
  });
}

function snapshotItemsMutationState(
  qc: QueryClient,
  ids: number[],
): ItemsMutationContext {
  return {
    prevItemLists: qc.getQueriesData<ItemsInfiniteData>({
      queryKey: queryKeys.items.lists(),
    }),
    prevItemDetails: ids.map(
      (id) =>
        [id, qc.getQueryData<Item>(queryKeys.items.detail(id))] as const,
    ),
    prevFeeds: qc.getQueryData<Feed[]>(queryKeys.feeds.list()),
    prevBookmarkLists: qc.getQueriesData<BookmarksInfiniteData>({
      queryKey: queryKeys.bookmarks.lists(),
    }),
  };
}

function applyOptimisticItemReadState(
  qc: QueryClient,
  ids: number[],
  targetUnread: boolean,
  prevFeeds: Feed[] | undefined,
) {
  const idSet = new Set(ids);
  const feedDeltaMap = new Map<number, number>();
  const updatedItemsById = new Map<number, Item>();

  const applyDelta = (item: Item): Item => {
    const delta = targetUnread ? 1 : -1;
    feedDeltaMap.set(
      item.feed_id,
      (feedDeltaMap.get(item.feed_id) ?? 0) + delta,
    );
    const updatedItem = { ...item, unread: targetUnread };
    updatedItemsById.set(item.id, updatedItem);
    return updatedItem;
  };

  // Item list caches carry the unread filter in their query key
  // (["items","list",{feedId,groupId,unread},pageSize]). When marking items
  // read inside an unread-filtered cache, those items stop matching the
  // filter, so they must be dropped and `total` shrunk — otherwise the
  // count-based getNextPageParam keeps a stale `data.length` that diverges
  // from the server's shrinking unread `total`, making the "load more" button
  // vanish while unread items still exist. Mirrors useDeleteBookmark.
  const itemLists = qc.getQueriesData<ItemsInfiniteData>({
    queryKey: queryKeys.items.lists(),
  });

  for (const [key, old] of itemLists) {
    if (!old) continue;

    const filters = key[2] as NormalizedItemFilters | undefined;
    const dropMatching = filters?.unread === true && !targetUnread;

    const pages = old.pages.map((page) => {
      if (dropMatching) {
        let removed = 0;
        const data = page.data.filter((item) => {
          if (!idSet.has(item.id) || item.unread === targetUnread) return true;
          removed += 1;
          applyDelta(item);
          return false;
        });
        if (removed === 0) return page;
        return { ...page, data, total: Math.max(0, page.total - removed) };
      }

      let changed = false;
      const data = page.data.map((item) => {
        if (!idSet.has(item.id) || item.unread === targetUnread) return item;
        changed = true;
        return applyDelta(item);
      });
      return changed ? { ...page, data } : page;
    });

    qc.setQueryData(key, { ...old, pages });
  }

  for (const id of ids) {
    const optimisticItem = updatedItemsById.get(id);
    qc.setQueryData<Item>(queryKeys.items.detail(id), (old) =>
      old
        ? old.unread !== targetUnread
          ? { ...old, unread: targetUnread }
          : old
        : optimisticItem,
    );
  }

  if (prevFeeds && feedDeltaMap.size > 0) {
    qc.setQueryData(queryKeys.feeds.list(), (old: Feed[] | undefined) =>
      old?.map((feed) => {
        const delta = feedDeltaMap.get(feed.id) ?? 0;
        if (delta === 0) return feed;

        return {
          ...feed,
          unread_count: Math.max(0, feed.unread_count + delta),
        };
      }),
    );
  }

  // Mirror the read-state change into bookmark caches so the starred view (which
  // renders bookmarks as articles) reflects the toggle without a refetch.
  qc.setQueriesData<BookmarksInfiniteData>(
    { queryKey: queryKeys.bookmarks.lists() },
    (old) => {
      if (!old) return old;
      let changed = false;
      const pages = old.pages.map((page) => {
        let pageChanged = false;
        const newData = page.data.map((bookmark) => {
          if (
            bookmark.item_id != null &&
            idSet.has(bookmark.item_id) &&
            bookmark.unread !== targetUnread
          ) {
            pageChanged = true;
            return { ...bookmark, unread: targetUnread };
          }
          return bookmark;
        });
        if (!pageChanged) return page;
        changed = true;
        return { ...page, data: newData };
      });
      return changed ? { ...old, pages } : old;
    },
  );
}

function rollbackItemsMutation(
  qc: QueryClient,
  context: ItemsMutationContext | undefined,
) {
  if (!context) return;

  for (const [key, data] of context.prevItemLists) {
    qc.setQueryData(key, data);
  }

  for (const [id, data] of context.prevItemDetails) {
    qc.setQueryData(queryKeys.items.detail(id), data);
  }

  if (context.prevFeeds) {
    qc.setQueryData(queryKeys.feeds.list(), context.prevFeeds);
  }

  for (const [key, data] of context.prevBookmarkLists) {
    qc.setQueryData(key, data);
  }
}

function useSetItemsReadState(targetUnread: boolean) {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: async (ids: number[]) => {
      if (targetUnread) {
        await itemAPI.markUnread({ ids });
      } else {
        await itemAPI.markRead({ ids });
      }

      return ids;
    },
    onMutate: async (ids) => {
      await Promise.all([
        qc.cancelQueries({ queryKey: queryKeys.items.all }),
        qc.cancelQueries({ queryKey: queryKeys.feeds.all }),
      ]);

      const context = snapshotItemsMutationState(qc, ids);
      applyOptimisticItemReadState(qc, ids, targetUnread, context.prevFeeds);
      return context;
    },
    onError: (_error, _ids, context) => {
      rollbackItemsMutation(qc, context);
    },
    onSettled: async () => {
      await qc.invalidateQueries({ queryKey: queryKeys.feeds.all });
    },
  });
}

export function useMarkItemsRead() {
  return useSetItemsReadState(false);
}

export function useMarkItemsUnread() {
  return useSetItemsReadState(true);
}
