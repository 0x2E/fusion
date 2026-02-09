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

const PAGE_SIZE = 10;

type ItemListResponse = Awaited<ReturnType<typeof itemAPI.list>>;
type ItemsInfiniteData = InfiniteData<ItemListResponse, number>;
type ItemsMutationContext = {
  prevItemLists: Array<[readonly unknown[], ItemsInfiniteData | undefined]>;
  prevItemDetails: Array<readonly [number, Item | undefined]>;
  prevFeeds: Feed[] | undefined;
};

function buildListItemsParams(
  filters: NormalizedItemFilters,
  offset: number,
): ListItemsParams {
  const params: ListItemsParams = {
    limit: PAGE_SIZE,
    offset,
    order_by: "pub_date:desc",
  };

  if (filters.feedId) params.feed_id = filters.feedId;
  if (filters.groupId) params.group_id = filters.groupId;
  if (filters.unread) params.unread = true;

  return params;
}

export const itemQueries = {
  list: (filters: ItemFilters) => {
    const normalizedFilters = normalizeItemFilters(filters);

    return infiniteQueryOptions({
      queryKey: queryKeys.items.list(normalizedFilters),
      queryFn: async ({ pageParam }) =>
        itemAPI.list(buildListItemsParams(normalizedFilters, pageParam)),
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

export function useItems(filters: ItemFilters) {
  return useInfiniteQuery(itemQueries.list(filters));
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

  qc.setQueriesData<ItemsInfiniteData>(
    { queryKey: queryKeys.items.lists() },
    (old) => {
      if (!old) return old;

      return {
        ...old,
        pages: old.pages.map((page) => ({
          ...page,
          data: page.data.map((item) => {
            if (!idSet.has(item.id) || item.unread === targetUnread) {
              return item;
            }

            const delta = targetUnread ? 1 : -1;
            feedDeltaMap.set(
              item.feed_id,
              (feedDeltaMap.get(item.feed_id) ?? 0) + delta,
            );

            return { ...item, unread: targetUnread };
          }),
        })),
      };
    },
  );

  for (const id of ids) {
    qc.setQueryData<Item>(queryKeys.items.detail(id), (old) =>
      old && old.unread !== targetUnread
        ? { ...old, unread: targetUnread }
        : old,
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
      await Promise.all([
        qc.invalidateQueries({ queryKey: queryKeys.items.all }),
        qc.invalidateQueries({ queryKey: queryKeys.feeds.all }),
      ]);
    },
  });
}

export function useMarkItemsRead() {
  return useSetItemsReadState(false);
}

export function useMarkItemsUnread() {
  return useSetItemsReadState(true);
}
