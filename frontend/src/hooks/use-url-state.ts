import {
  useNavigate,
  useParams,
  useSearch,
} from "@tanstack/react-router";
import { useCallback } from "react";
import {
  defaultArticleFilter,
  isArticleFilter,
  type ArticleFilter,
} from "@/lib/article-filter";
import { parsePositiveIntegerParam } from "@/lib/route-params";

export type { ArticleFilter } from "@/lib/article-filter";

export function useUrlState() {
  const navigate = useNavigate();
  const params = useParams({ strict: false }) as {
    filter?: string;
    feedId?: string;
    groupId?: string;
  };
  const search = useSearch({ strict: false }) as Record<string, unknown>;

  const routeFeedId = parsePositiveIntegerParam(params.feedId);
  const routeGroupId = parsePositiveIntegerParam(params.groupId);
  const selectedArticleId = parsePositiveIntegerParam(search.article);
  const routeFilter =
    typeof params.filter === "string" && isArticleFilter(params.filter)
      ? params.filter
      : defaultArticleFilter;

  const selectedFeedId = routeFeedId;
  const selectedGroupId = routeGroupId;
  const articleFilter = routeFilter;

  const navigateToList = useCallback(
    ({
      filter,
      feedId,
      groupId,
      articleId,
      replace,
    }: {
      filter?: ArticleFilter;
      feedId?: number | null;
      groupId?: number | null;
      articleId?: number | null;
      replace?: boolean;
    } = {}) => {
      const nextFilter = filter ?? articleFilter;
      const nextFeedId = feedId === undefined ? selectedFeedId : feedId;
      const nextGroupId = groupId === undefined ? selectedGroupId : groupId;
      const nextArticleId = articleId === undefined ? selectedArticleId : articleId;
      const nextSearch = { article: nextArticleId ?? undefined };

      if (nextGroupId !== null) {
        navigate({
          to: "/groups/$groupId/$filter",
          params: {
            groupId: String(nextGroupId),
            filter: nextFilter,
          },
          search: nextSearch,
          replace: replace ?? true,
        });
        return;
      }

      if (nextFeedId !== null) {
        navigate({
          to: "/feeds/$feedId/$filter",
          params: {
            feedId: String(nextFeedId),
            filter: nextFilter,
          },
          search: nextSearch,
          replace: replace ?? true,
        });
        return;
      }

      navigate({
        to: "/$filter",
        params: { filter: nextFilter },
        search: nextSearch,
        replace: replace ?? true,
      });
    },
    [
      articleFilter,
      navigate,
      selectedArticleId,
      selectedFeedId,
      selectedGroupId,
    ],
  );

  const setSelectedFeed = useCallback(
    (feedId: number | null) => {
      navigateToList({
        feedId,
        groupId: null,
        articleId: null,
      });
    },
    [navigateToList],
  );

  const setSelectedGroup = useCallback(
    (groupId: number | null) => {
      navigateToList({
        groupId,
        feedId: null,
        articleId: null,
      });
    },
    [navigateToList],
  );

  const setSelectedArticle = useCallback(
    (articleId: number | null) => {
      if (articleId === null) {
        navigateToList({ articleId: null, replace: true });
        return;
      }

      navigateToList({
        articleId,
        replace: selectedArticleId !== null,
      });
    },
    [navigateToList, selectedArticleId],
  );

  const setArticleFilter = useCallback(
    (filter: ArticleFilter) => {
      navigateToList({
        filter,
        articleId: null,
      });
    },
    [navigateToList],
  );

  const selectTopLevelFilter = useCallback(
    (filter: ArticleFilter) => {
      navigateToList({
        filter,
        feedId: null,
        groupId: null,
        articleId: null,
      });
    },
    [navigateToList],
  );

  const selectAll = useCallback(() => {
    selectTopLevelFilter("all");
  }, [selectTopLevelFilter]);

  return {
    selectedFeedId,
    selectedGroupId,
    selectedArticleId,
    articleFilter,
    setSelectedFeed,
    setSelectedGroup,
    setSelectedArticle,
    setArticleFilter,
    selectTopLevelFilter,
    selectAll,
  };
}
