import { useCallback, useMemo, useRef, useState } from "react";
import { type InfiniteData, useQueryClient } from "@tanstack/react-query";
import type { Bookmark, Item } from "@/lib/api";
import {
  getArticleListContextKey,
  type ArticleListContext,
} from "@/lib/article-list-context";
import {
  createEmptyArticleListOverlay,
  useArticleSessionStore,
} from "@/store/article-session";
import { usePreferencesStore } from "@/store";
import {
  bookmarkQueries,
  fetchBookmarkPage,
  type BookmarkPage,
  useBookmarkLookup,
  useStarredItems,
} from "./bookmarks";
import {
  countStickyRemovedRows,
  getAdjustedOffset,
  mergeStickyArticles,
} from "./article-list-logic";
import { queryKeys } from "./keys";
import { fetchItemsPage, itemQueries, type ItemListResponse, useItems } from "./items";

export interface ArticleListData {
  articles: Item[];
  hasMore: boolean | undefined;
  isLoading: boolean;
  isLoadingMore: boolean;
  loadMore: () => Promise<unknown>;
  isItemStarred: (itemId: number) => boolean;
  getBookmarkByItemId: (itemId: number) => Bookmark | undefined;
}

function applyArticleOverrides(
  article: Item,
  readOverrides: Record<number, boolean>,
): Item {
  const unread = readOverrides[article.id];
  return unread === undefined ? article : { ...article, unread };
}

export function useArticleListData(context: ArticleListContext): ArticleListData {
  const queryClient = useQueryClient();
  const articlePageSize = usePreferencesStore((state) => state.articlePageSize);
  const inFlightOffsetsRef = useRef(new Set<string>());
  const [isManualLoadingMore, setIsManualLoadingMore] = useState(false);
  const contextKey = getArticleListContextKey(context);
  const overlay = useArticleSessionStore(
    (state) => state.overlays[contextKey] ?? createEmptyArticleListOverlay(),
  );
  const itemsQuery = useItems({
    feedId: context.feedId,
    groupId: context.groupId,
    unread: context.filter === "unread" ? true : undefined,
  });
  const starredItems = useStarredItems({
    feedId: context.feedId,
    groupId: context.groupId,
  });
  const {
    fetchedCount: fetchedBookmarkCount,
    total: bookmarkTotal,
    getBookmarkByItemId,
    isItemStarred: isBaselineItemStarred,
    isLoading: isBookmarksLoading,
  } = useBookmarkLookup();
  const itemPages = itemsQuery.data?.pages ?? [];
  const unreadTotal = itemPages.at(-1)?.total ?? 0;
  const unreadLoadedCount = itemPages.reduce((count, page) => count + page.data.length, 0);
  const unreadRemovedCount = itemPages.reduce(
    (count, page) =>
      count + page.data.filter((article) => overlay.readOverrides[article.id] === false).length,
    0,
  );
  const starredLoadedCount = fetchedBookmarkCount;
  const starredRemovedCount = countStickyRemovedRows(
    overlay.stickyVisibleIds,
    overlay.starOverrides,
  );
  const adjustedUnreadOffset = getAdjustedOffset(
    unreadLoadedCount,
    unreadRemovedCount,
  );
  const adjustedStarredOffset = getAdjustedOffset(
    starredLoadedCount,
    starredRemovedCount,
  );

  const articles = useMemo(() => {
    const baselineArticles =
      context.filter === "starred"
        ? starredItems.articles
        : itemsQuery.data?.pages.flatMap((page) => page.data) ?? [];
    const visibleArticles = mergeStickyArticles({
      baselineArticles,
      stickyVisibleIds: overlay.stickyVisibleIds,
      stickyArticles: overlay.stickyArticles,
      getArticleId: (article) => article.id,
      getCachedArticle: (itemId) => {
        if (itemId <= 0) {
          return undefined;
        }

        return queryClient.getQueryData<Item>(queryKeys.items.detail(itemId));
      },
    }).map((article) => {
      if (context.filter !== "starred" || article.id <= 0) {
        return article;
      }

      const cachedItem = queryClient.getQueryData<Item>(
        queryKeys.items.detail(article.id),
      );
      return cachedItem ?? article;
    });

    return visibleArticles
      .filter((article) => {
        if (context.filter === "all") {
          return true;
        }

        if (context.filter === "unread") {
          const unread = overlay.readOverrides[article.id] ?? article.unread;
          return unread || overlay.stickyVisibleIds[article.id] === true;
        }

        const starred = overlay.starOverrides[article.id] ?? true;
        return starred || overlay.stickyVisibleIds[article.id] === true;
      })
      .map((article) => applyArticleOverrides(article, overlay.readOverrides));
  }, [
    context.filter,
    itemsQuery.data,
    overlay.readOverrides,
    overlay.starOverrides,
    overlay.stickyArticles,
    overlay.stickyVisibleIds,
    queryClient,
    starredItems.articles,
  ]);

  const isItemStarred = useCallback(
    (itemId: number) => overlay.starOverrides[itemId] ?? isBaselineItemStarred(itemId),
    [isBaselineItemStarred, overlay.starOverrides],
  );

  const appendItemsPage = useCallback(
    async (offset: number) => {
      const requestKey = `unread:${offset}`;
      if (inFlightOffsetsRef.current.has(requestKey)) {
        return undefined;
      }

      const query = itemQueries.list(
        {
          feedId: context.feedId,
          groupId: context.groupId,
          unread: true,
        },
        articlePageSize,
      );

      inFlightOffsetsRef.current.add(requestKey);
      setIsManualLoadingMore(true);

      try {
        const page = await fetchItemsPage(
          {
            feedId: context.feedId,
            groupId: context.groupId,
            unread: true,
          },
          articlePageSize,
          offset,
        );

        queryClient.setQueryData<InfiniteData<ItemListResponse, number>>(
          query.queryKey,
          (old) => {
            if (!old) {
              return {
                pages: [page],
                pageParams: [offset],
              };
            }

            const existingIndex = old.pageParams.indexOf(offset);
            if (existingIndex !== -1) {
              const nextPages = [...old.pages];
              nextPages[existingIndex] = page;

              return {
                ...old,
                pages: nextPages,
              };
            }

            return {
              ...old,
              pages: [...old.pages, page],
              pageParams: [...old.pageParams, offset],
            };
          },
        );

        return page;
      } finally {
        inFlightOffsetsRef.current.delete(requestKey);
        setIsManualLoadingMore(inFlightOffsetsRef.current.size > 0);
      }
    },
    [articlePageSize, context.feedId, context.groupId, queryClient],
  );

  const appendBookmarkPage = useCallback(
    async (offset: number) => {
      const requestKey = `starred:${offset}`;
      if (inFlightOffsetsRef.current.has(requestKey)) {
        return undefined;
      }

      const queryKey = bookmarkQueries.list().queryKey;
      inFlightOffsetsRef.current.add(requestKey);
      setIsManualLoadingMore(true);

      try {
        const page = await fetchBookmarkPage(offset);

        queryClient.setQueryData<InfiniteData<BookmarkPage, number>>(
          queryKey,
          (old) => {
            if (!old) {
              return {
                pages: [page],
                pageParams: [offset],
              };
            }

            const existingIndex = old.pageParams.indexOf(offset);
            if (existingIndex !== -1) {
              const nextPages = [...old.pages];
              nextPages[existingIndex] = page;

              return {
                ...old,
                pages: nextPages,
              };
            }

            return {
              ...old,
              pages: [...old.pages, page],
              pageParams: [...old.pageParams, offset],
            };
          },
        );

        return page;
      } finally {
        inFlightOffsetsRef.current.delete(requestKey);
        setIsManualLoadingMore(inFlightOffsetsRef.current.size > 0);
      }
    },
    [queryClient],
  );

  return {
    articles,
    hasMore:
      context.filter === "unread"
        ? adjustedUnreadOffset < unreadTotal
        : context.filter === "starred"
          ? adjustedStarredOffset < bookmarkTotal
          : itemsQuery.hasNextPage,
    isLoading:
      context.filter === "starred" ? isBookmarksLoading : itemsQuery.isLoading,
    isLoadingMore:
      context.filter === "unread" || context.filter === "starred"
        ? isManualLoadingMore
        : itemsQuery.isFetchingNextPage,
    loadMore: () => {
      if (context.filter === "unread") {
        return appendItemsPage(adjustedUnreadOffset);
      }

      if (context.filter === "starred") {
        return appendBookmarkPage(adjustedStarredOffset);
      }

      return itemsQuery.fetchNextPage();
    },
    isItemStarred,
    getBookmarkByItemId: (itemId: number): Bookmark | undefined =>
      getBookmarkByItemId(itemId),
  };
}
