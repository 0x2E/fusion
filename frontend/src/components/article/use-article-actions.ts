import { useCallback } from "react";
import { useQueryClient } from "@tanstack/react-query";
import {
  getArticleListContextKey,
  type ArticleListContext,
} from "@/lib/article-list-context";
import type { Item } from "@/lib/api";
import { useCreateBookmark, useDeleteBookmark } from "@/queries/bookmarks";
import type { ArticleListData } from "@/queries/article-list";
import {
  itemQueries,
  useMarkItemsRead,
  useMarkItemsUnread,
} from "@/queries/items";
import { useArticleSessionStore } from "@/store/article-session";

export function useArticleActions(
  context: ArticleListContext,
  articleList: ArticleListData,
) {
  const queryClient = useQueryClient();
  const contextKey = getArticleListContextKey(context);
  const overlay = useArticleSessionStore((state) => state.overlays[contextKey]);
  const keepVisible = useArticleSessionStore((state) => state.keepVisible);
  const setStickyArticle = useArticleSessionStore((state) => state.setStickyArticle);
  const setReadOverride = useArticleSessionStore((state) => state.setReadOverride);
  const setStarOverride = useArticleSessionStore((state) => state.setStarOverride);
  const setOverlay = useArticleSessionStore((state) => state.setOverlay);
  const markRead = useMarkItemsRead();
  const markUnread = useMarkItemsUnread();
  const createBookmark = useCreateBookmark();
  const deleteBookmark = useDeleteBookmark();
  const isStarredMode = context.filter === "starred";

  const toggleRead = useCallback(
    async (article: Item) => {
      if (article.id <= 0) {
        return;
      }

      const previousOverlay = overlay;
      let unread = article.unread;

      if (isStarredMode) {
        try {
          const detail = await queryClient.ensureQueryData(itemQueries.detail(article.id));
          if (detail === undefined) {
            return;
          }

          unread = detail.unread;
        } catch {
          return;
        }
      }

      try {
        setReadOverride(contextKey, article.id, !unread);

        if (context.filter === "unread" && unread) {
          keepVisible(contextKey, article.id);
        }

        if (unread) {
          await markRead.mutateAsync([article.id]);
        } else {
          await markUnread.mutateAsync([article.id]);
        }
      } catch (error) {
        setOverlay(contextKey, previousOverlay);
        throw error;
      }
    },
    [
      context.filter,
      contextKey,
      isStarredMode,
      keepVisible,
      markRead,
      markUnread,
      overlay,
      queryClient,
      setOverlay,
      setReadOverride,
    ],
  );

  const toggleStar = useCallback(
    async (article: Item) => {
      const previousOverlay = overlay;
      const starred = articleList.isItemStarred(article.id);

      try {
        if (context.filter === "starred" && starred) {
          keepVisible(contextKey, article.id);
          setStickyArticle(contextKey, article);
          setStarOverride(contextKey, article.id, false);
        } else {
          setStarOverride(contextKey, article.id, !starred);
        }

        if (starred) {
          const bookmarkId = articleList.getSavedBookmarkIdByItemId(article.id);
          if (bookmarkId !== undefined) {
            await deleteBookmark.mutateAsync(bookmarkId);
          }
          return;
        }

        await createBookmark.mutateAsync(article);
      } catch (error) {
        setOverlay(contextKey, previousOverlay);
        throw error;
      }
    },
    [
      articleList,
      context.filter,
      contextKey,
      createBookmark,
      deleteBookmark,
      keepVisible,
      overlay,
      setStickyArticle,
      setOverlay,
      setStarOverride,
    ],
  );

  return { toggleRead, toggleStar };
}
