import { useCallback, useEffect } from "react";
import { useNavigate } from "@tanstack/react-router";
import { useQueryClient } from "@tanstack/react-query";
import { CheckCheck, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ArticleItem } from "./article-item";
import { ContentHeader } from "@/components/layout/content-header";
import { SidebarTrigger } from "@/components/layout/sidebar-trigger";
import { useArticleNavigation } from "@/hooks/use-keyboard";
import { useUrlState, type ArticleFilter } from "@/hooks/use-url-state";
import { getArticleListContextKey } from "@/lib/article-list-context";
import { itemQueries, useMarkItemsRead } from "@/queries/items";
import { useFeedLookup } from "@/queries/feeds";
import { useGroups } from "@/queries/groups";
import { useArticleListData } from "@/queries/article-list";
import { getFaviconUrl } from "@/lib/api/favicon";
import { useI18n } from "@/lib/i18n";
import { useArticleSessionStore } from "@/store/article-session";
import { useArticleActions } from "./use-article-actions";

export function ArticleList() {
  const { t } = useI18n();
  const navigate = useNavigate();
  const {
    articleFilter,
    articleListContext,
    setArticleFilter,
    selectedArticleId,
    setSelectedArticle,
  } = useUrlState();
  const queryClient = useQueryClient();
  const contextKey = getArticleListContextKey(articleListContext);
  const articleList = useArticleListData(articleListContext);
  const overlay = useArticleSessionStore(
    (state) => state.overlays[contextKey],
  );
  const keepVisible = useArticleSessionStore((state) => state.keepVisible);
  const setReadOverride = useArticleSessionStore((state) => state.setReadOverride);
  const setOverlay = useArticleSessionStore((state) => state.setOverlay);
  const clearContext = useArticleSessionStore((state) => state.clearContext);

  const { data: groups = [] } = useGroups();
  const { feeds, getFeedById, isLoading: isFeedsLoading } = useFeedLookup();
  const markItemsRead = useMarkItemsRead();
  const articles = articleList.articles;
  const isStarredMode = articleFilter === "starred";
  const { toggleRead, toggleStar } = useArticleActions(articleListContext, articleList);

  useEffect(() => {
    return () => {
      clearContext(contextKey);
    };
  }, [clearContext, contextKey]);

  // Setup keyboard navigation
  const articleIds = articles.map((a) => a.id);
  useArticleNavigation(articleIds, {
    enabled: selectedArticleId === null,
  });

  // Determine title
  let title = t("article.list.all");
  if (articleListContext.feedId) {
    const feed = getFeedById(articleListContext.feedId);
    title = feed?.name ?? t("article.feedFallback");
  } else if (articleListContext.groupId) {
    const group = groups.find((g) => g.id === articleListContext.groupId);
    title = group?.name ?? t("article.groupFallback");
  }

  const unreadCount = articles.filter((a) => a.unread).length;
  const hasNoFeeds = !isFeedsLoading && feeds.length === 0;

  const handleToggleRead = useCallback(
    async (article: (typeof articles)[number]) => {
      try {
        await toggleRead(article);
      } catch (error) {
        console.error("Failed to toggle read status:", error);
      }
    },
    [toggleRead],
  );

  const handleToggleStar = useCallback(
    async (article: (typeof articles)[number]) => {
      try {
        await toggleStar(article);
      } catch (error) {
        console.error("Failed to toggle star:", error);
      }
    },
    [toggleStar],
  );

  const handleMarkAllAsRead = async () => {
    const previousOverlay = overlay;
    let unreadIds = articles
      .filter((a) => a.unread && a.id > 0)
      .map((a) => a.id);

    if (isStarredMode) {
      const ids = articles.filter((a) => a.id > 0).map((a) => a.id);
      const detailEntries = await Promise.all(
        ids.map(async (id) => {
          try {
            const detail = await queryClient.ensureQueryData(
              itemQueries.detail(id),
            );
            return [id, detail?.unread ?? false] as const;
          } catch {
            return [id, false] as const;
          }
        }),
      );

      unreadIds = detailEntries
        .filter(([, unread]) => unread)
        .map(([id]) => id);
    }

    if (unreadIds.length === 0) return;

    try {
      if (articleFilter === "unread") {
        for (const id of unreadIds) {
          keepVisible(contextKey, id);
          setReadOverride(contextKey, id, false);
        }
      } else {
        for (const id of unreadIds) {
          setReadOverride(contextKey, id, false);
        }
      }

      await markItemsRead.mutateAsync(unreadIds);
    } catch (error) {
      setOverlay(contextKey, previousOverlay);
      console.error("Failed to mark all as read:", error);
    }
  };

  return (
    <div className="flex h-full flex-col">
      <ContentHeader>
        <div className="flex min-w-0 items-center gap-1">
          <SidebarTrigger />
          <h2 className="truncate text-lg font-semibold">{title}</h2>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={handleMarkAllAsRead}
          disabled={unreadCount === 0}
          className="gap-1.5 text-xs"
        >
          <CheckCheck className="h-4 w-4" />
          {t("article.list.markAllRead")}
        </Button>
      </ContentHeader>

      {/* Article area with filter tabs */}
      <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden px-4 py-4 sm:px-6">
        {/* Filter tabs - hidden when no articles exist */}
        {!hasNoFeeds && (articles.length > 0 || articleFilter !== "all") && (
          <Tabs
            value={articleFilter}
            onValueChange={(v) => setArticleFilter(v as ArticleFilter)}
          >
            <TabsList>
              <TabsTrigger value="all">{t("article.filter.all")}</TabsTrigger>
              <TabsTrigger value="unread">{t("article.filter.unread")}</TabsTrigger>
              <TabsTrigger value="starred">
                {t("article.filter.starred")}
              </TabsTrigger>
            </TabsList>
          </Tabs>
        )}

        {/* Article list */}
        <ScrollArea className="min-h-0 flex-1">
          <div>
            {articleList.isLoading && articles.length === 0 ? (
              <div className="space-y-2 p-2">
                {[1, 2, 3, 4, 5].map((i) => (
                  <div
                    key={i}
                    className="h-24 animate-pulse rounded-md bg-accent"
                  />
                ))}
              </div>
            ) : articles.length === 0 ? (
              hasNoFeeds ? (
                <div className="flex flex-col items-center justify-center gap-3 py-12 text-center">
                  <p className="text-sm text-muted-foreground">
                    {t("article.list.noFeeds")}
                  </p>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => navigate({ to: "/feeds" })}
                  >
                    {t("article.list.openFeedManagement")}
                  </Button>
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <p className="text-sm text-muted-foreground">
                    {t("article.list.noArticles")}
                  </p>
                </div>
              )
            ) : (
              <>
                {articles.map((article) => {
                  const feed = getFeedById(article.feed_id);
                  const bookmark = articleList.getBookmarkByItemId(article.id);

                  return (
                    <ArticleItem
                      key={article.id}
                      article={article}
                      selectedArticleId={selectedArticleId}
                      onSelectArticle={setSelectedArticle}
                      onToggleRead={handleToggleRead}
                      onToggleStar={handleToggleStar}
                      canToggleRead={article.id > 0}
                      isStarred={articleList.isItemStarred(article.id)}
                      feedName={feed?.name ?? bookmark?.feed_name ?? t("common.unknown")}
                      feedFaviconUrl={
                        feed ? getFaviconUrl(feed.link, feed.site_url) : null
                      }
                    />
                  );
                })}
                {articleList.hasMore && (
                  <div className="flex justify-center py-4">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        void articleList.loadMore();
                      }}
                      disabled={articleList.isLoadingMore}
                      className="gap-2"
                    >
                      {articleList.isLoadingMore && (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      )}
                      {articleList.isLoadingMore
                        ? t("article.list.loading")
                        : t("article.list.loadMore")}
                    </Button>
                  </div>
                )}
              </>
            )}
          </div>
        </ScrollArea>
      </div>
    </div>
  );
}
