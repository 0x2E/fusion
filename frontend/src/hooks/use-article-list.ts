import { useCallback, useMemo } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useItems } from "@/queries/items";
import { useStarredItems } from "@/queries/bookmarks";
import { queryKeys } from "@/queries/keys";
import type { Item } from "@/lib/api";
import type { ArticleFilter } from "@/lib/article-filter";

interface ArticleListFilters {
  feedId: number | null;
  groupId: number | null;
  articleFilter: ArticleFilter;
}

export function useArticleList(filters: ArticleListFilters) {
  const queryClient = useQueryClient();
  const isStarredMode = filters.articleFilter === "starred";

  const itemsQuery = useItems({
    feedId: filters.feedId,
    groupId: filters.groupId,
    unread: filters.articleFilter === "unread" ? true : undefined,
  });

  const items = useMemo(
    () => itemsQuery.data?.pages.flatMap((p) => p.data) ?? [],
    [itemsQuery.data],
  );

  const starredItems = useStarredItems({
    feedId: filters.feedId,
    groupId: filters.groupId,
  });

  const articles = isStarredMode ? starredItems : items;

  const getArticleUnread = useCallback(
    (article: Item) => {
      if (!isStarredMode) return article.unread;

      if (article.id > 0) {
        const cachedItem = queryClient.getQueryData<Item>(
          queryKeys.items.detail(article.id),
        );
        if (cachedItem) return cachedItem.unread;
      }

      return article.unread;
    },
    [isStarredMode, queryClient],
  );

  const displayArticles = useMemo(
    () =>
      articles.map((article) => ({
        ...article,
        unread: getArticleUnread(article),
      })),
    [articles, getArticleUnread],
  );

  return {
    articles,
    displayArticles,
    hasMore: isStarredMode ? false : itemsQuery.hasNextPage,
    isLoading: isStarredMode ? false : itemsQuery.isLoading,
    isLoadingMore: itemsQuery.isFetchingNextPage,
    isStarredMode,
    fetchNextPage: itemsQuery.fetchNextPage,
  };
}
