import { useNavigate, useSearch } from "@tanstack/react-router";
import { useCallback } from "react";
import type { SearchParams } from "@/routes/index";

export type ArticleFilter = "all" | "unread" | "starred";

export function useUrlState() {
  const navigate = useNavigate();
  const search = useSearch({ from: "/" });

  const selectedFeedId = search.feed ?? null;
  const selectedGroupId = search.group ?? null;
  const selectedArticleId = search.article ?? null;
  const articleFilter: ArticleFilter = search.filter ?? "all";

  const setSelectedFeed = useCallback(
    (feedId: number | null) => {
      navigate({
        to: "/",
        search: (prev: SearchParams) => ({
          ...prev,
          feed: feedId ?? undefined,
          group: undefined,
        }),
        replace: true,
      });
    },
    [navigate]
  );

  const setSelectedGroup = useCallback(
    (groupId: number | null) => {
      navigate({
        to: "/",
        search: (prev: SearchParams) => ({
          ...prev,
          group: groupId ?? undefined,
          feed: undefined,
        }),
        replace: true,
      });
    },
    [navigate]
  );

  const setSelectedArticle = useCallback(
    (articleId: number | null) => {
      navigate({
        to: "/",
        search: (prev: SearchParams) => ({
          ...prev,
          article: articleId ?? undefined,
        }),
        replace: true,
      });
    },
    [navigate]
  );

  const setArticleFilter = useCallback(
    (filter: ArticleFilter) => {
      navigate({
        to: "/",
        search: (prev: SearchParams) => ({
          ...prev,
          filter: filter === "all" ? undefined : filter,
        }),
        replace: true,
      });
    },
    [navigate]
  );

  const selectAll = useCallback(() => {
    navigate({
      to: "/",
      search: (prev: SearchParams) => ({
        ...prev,
        feed: undefined,
        group: undefined,
      }),
      replace: true,
    });
  }, [navigate]);

  return {
    selectedFeedId,
    selectedGroupId,
    selectedArticleId,
    articleFilter,
    setSelectedFeed,
    setSelectedGroup,
    setSelectedArticle,
    setArticleFilter,
    selectAll,
  };
}
