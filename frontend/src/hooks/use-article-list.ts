import { useCallback, useMemo } from "react";
import { useItems } from "@/queries/items";
import {
  resolveBookmarkItemId,
  useBookmarkLookup,
  useStarredItems,
} from "@/queries/bookmarks";
import type { Bookmark, Item } from "@/lib/api";
import type { ArticleFilter } from "@/lib/article-filter";

interface ArticleListFilters {
  feedId: number | null;
  groupId: number | null;
  articleFilter: ArticleFilter;
}

export function useArticleList(filters: ArticleListFilters) {
  const isStarredMode = filters.articleFilter === "starred";

  // Items are unused in starred mode (bookmarks ARE the articles there), so
  // skip the request entirely.
  const itemsQuery = useItems(
    {
      feedId: filters.feedId,
      groupId: filters.groupId,
      unread: filters.articleFilter === "unread" ? true : undefined,
    },
    !isStarredMode,
  );

  // Unfiltered lookup: powers star indicators in non-starred views.
  const lookup = useBookmarkLookup();

  // Server-filtered, paginated starred list. Only fetched in starred mode.
  const starred = useStarredItems(
    { feedId: filters.feedId, groupId: filters.groupId },
    isStarredMode,
  );

  const items = useMemo(
    () => itemsQuery.data?.pages.flatMap((p) => p.data) ?? [],
    [itemsQuery.data],
  );

  const articles: Item[] = isStarredMode ? starred.items : items;

  // Bookmark resolution for star state + un-starring. In starred mode every
  // displayed article is itself a bookmark, so resolve from the starred data
  // (which may exceed the lookup's first page). Otherwise use the lookup.
  const bookmarkSource: Bookmark[] = isStarredMode
    ? starred.bookmarks
    : lookup.bookmarks;
  const bookmarkByItemId = useMemo(
    () =>
      new Map(bookmarkSource.map((b) => [resolveBookmarkItemId(b), b])),
    [bookmarkSource],
  );

  const isItemStarred = useCallback(
    (itemId: number) =>
      isStarredMode ? true : bookmarkByItemId.has(itemId),
    [isStarredMode, bookmarkByItemId],
  );

  const getBookmarkByItemId = useCallback(
    (itemId: number) => bookmarkByItemId.get(itemId),
    [bookmarkByItemId],
  );

  return {
    articles,
    hasMore: isStarredMode ? starred.hasNextPage : itemsQuery.hasNextPage,
    isLoading: isStarredMode ? starred.isLoading : itemsQuery.isLoading,
    isLoadingMore: isStarredMode
      ? starred.isFetchingNextPage
      : itemsQuery.isFetchingNextPage,
    isStarredMode,
    fetchNextPage: isStarredMode
      ? starred.fetchNextPage
      : () => void itemsQuery.fetchNextPage(),
    isItemStarred,
    getBookmarkByItemId,
  };
}
