import { useCallback, useMemo, useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { useQueryClient } from "@tanstack/react-query";
import { CheckCheck, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ArticleItem } from "./article-item";
import { SidebarTrigger } from "@/components/layout/sidebar-trigger";
import { useArticleNavigation } from "@/hooks/use-keyboard";
import { useUrlState, type ArticleFilter } from "@/hooks/use-url-state";
import {
  itemQueries,
  useItems,
  useMarkItemsRead,
  useMarkItemsUnread,
} from "@/queries/items";
import { useFeedLookup } from "@/queries/feeds";
import { useGroups } from "@/queries/groups";
import {
  useBookmarkLookup,
  useCreateBookmark,
  useDeleteBookmark,
  useStarredItems,
} from "@/queries/bookmarks";
import { queryKeys } from "@/queries/keys";
import { getFaviconUrl } from "@/lib/api/favicon";
import type { Item } from "@/lib/api";

export function ArticleList() {
  const navigate = useNavigate();
  const {
    articleFilter,
    setArticleFilter,
    selectedFeedId,
    selectedGroupId,
    selectedArticleId,
    setSelectedArticle,
  } = useUrlState();
  const queryClient = useQueryClient();
  const [starredUnreadOverrides, setStarredUnreadOverrides] = useState<
    Record<number, boolean>
  >({});

  const isStarredMode = articleFilter === "starred";

  // Items query for non-starred modes
  const itemsQuery = useItems({
    feedId: selectedFeedId,
    groupId: selectedGroupId,
    unread: articleFilter === "unread" ? true : undefined,
  });

  const { data: groups = [] } = useGroups();
  const { feeds, getFeedById, isLoading: isFeedsLoading } = useFeedLookup();
  const markItemsRead = useMarkItemsRead();
  const markItemsUnread = useMarkItemsUnread();
  const { isItemStarred, getBookmarkByItemId } = useBookmarkLookup();
  const createBookmark = useCreateBookmark();
  const deleteBookmark = useDeleteBookmark();

  // Flatten infinite query pages
  const items = useMemo(
    () => itemsQuery.data?.pages.flatMap((p) => p.data) ?? [],
    [itemsQuery.data],
  );

  const starredArticles = useStarredItems({
    feedId: selectedFeedId,
    groupId: selectedGroupId,
  });

  const articles = isStarredMode ? starredArticles : items;
  const getArticleUnread = useCallback(
    (article: Item) => {
      if (!isStarredMode) return article.unread;

      const override = starredUnreadOverrides[article.id];
      if (override !== undefined) return override;

      if (article.id > 0) {
        const cachedItem = queryClient.getQueryData<Item>(
          queryKeys.items.detail(article.id),
        );
        if (cachedItem) return cachedItem.unread;
      }

      return article.unread;
    },
    [isStarredMode, queryClient, starredUnreadOverrides],
  );

  const displayArticles = useMemo(
    () =>
      articles.map((article) => ({
        ...article,
        unread: getArticleUnread(article),
      })),
    [articles, getArticleUnread],
  );

  const hasMore = isStarredMode ? false : itemsQuery.hasNextPage;
  const isLoading = isStarredMode ? false : itemsQuery.isLoading;
  const isLoadingMore = itemsQuery.isFetchingNextPage;

  // Setup keyboard navigation
  const articleIds = displayArticles.map((a) => a.id);
  useArticleNavigation(articleIds);

  // Determine title
  let title = "All Articles";
  if (selectedFeedId) {
    const feed = getFeedById(selectedFeedId);
    title = feed?.name ?? "Feed";
  } else if (selectedGroupId) {
    const group = groups.find((g) => g.id === selectedGroupId);
    title = group?.name ?? "Group";
  }

  const unreadCount = displayArticles.filter((a) => a.unread).length;
  const hasNoFeeds = !isFeedsLoading && feeds.length === 0;

  const handleToggleRead = useCallback(
    async (article: Item) => {
      if (isStarredMode && article.id <= 0) {
        return;
      }

      let unread = getArticleUnread(article);

      if (isStarredMode && article.id > 0) {
        try {
          const detail = await queryClient.ensureQueryData(
            itemQueries.detail(article.id),
          );
          if (detail === undefined) {
            return;
          }

          unread = detail.unread;
        } catch {
          return;
        }
      }

      try {
        if (unread) {
          await markItemsRead.mutateAsync([article.id]);
        } else {
          await markItemsUnread.mutateAsync([article.id]);
        }

        if (isStarredMode) {
          setStarredUnreadOverrides((prev) => ({
            ...prev,
            [article.id]: !unread,
          }));
        }
      } catch (error) {
        console.error("Failed to toggle read status:", error);
      }
    },
    [
      getArticleUnread,
      isStarredMode,
      markItemsRead,
      markItemsUnread,
      queryClient,
    ],
  );

  const handleToggleStar = useCallback(
    async (article: Item) => {
      try {
        if (isItemStarred(article.id)) {
          const bookmark = getBookmarkByItemId(article.id);
          if (bookmark) {
            await deleteBookmark.mutateAsync(bookmark.id);
          }
          return;
        }

        await createBookmark.mutateAsync(article);
      } catch (error) {
        console.error("Failed to toggle star:", error);
      }
    },
    [createBookmark, deleteBookmark, getBookmarkByItemId, isItemStarred],
  );

  const handleMarkAllAsRead = async () => {
    let unreadIds = displayArticles
      .filter((a) => a.unread && a.id > 0)
      .map((a) => a.id);

    if (isStarredMode) {
      const ids = displayArticles.filter((a) => a.id > 0).map((a) => a.id);
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
      await markItemsRead.mutateAsync(unreadIds);

      if (isStarredMode) {
        setStarredUnreadOverrides((prev) => {
          const next = { ...prev };
          for (const id of unreadIds) {
            next[id] = false;
          }
          return next;
        });
      }
    } catch (error) {
      console.error("Failed to mark all as read:", error);
    }
  };

  return (
    <div className="flex h-full flex-col">
      {/* Header */}
      <div className="flex items-center justify-between border-b px-4 py-3 sm:px-6">
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
          Mark all as read
        </Button>
      </div>

      {/* Article area with filter tabs */}
      <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden px-4 py-4 sm:px-6">
        {/* Filter tabs - hidden when no articles exist */}
        {!hasNoFeeds && (articles.length > 0 || articleFilter !== "all") && (
          <Tabs
            value={articleFilter}
            onValueChange={(v) => setArticleFilter(v as ArticleFilter)}
          >
            <TabsList>
              <TabsTrigger value="all">All</TabsTrigger>
              <TabsTrigger value="unread">Unread</TabsTrigger>
              <TabsTrigger value="starred">Starred</TabsTrigger>
            </TabsList>
          </Tabs>
        )}

        {/* Article list */}
        <ScrollArea className="min-h-0 flex-1">
          <div>
            {isLoading && articles.length === 0 ? (
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
                    No feeds yet. Go to Feed Management to add your first feed.
                  </p>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => navigate({ to: "/feeds" })}
                  >
                    Open Feed Management
                  </Button>
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <p className="text-sm text-muted-foreground">
                    No articles found
                  </p>
                </div>
              )
            ) : (
              <>
                {displayArticles.map((article) => {
                  const feed = getFeedById(article.feed_id);
                  const bookmark = getBookmarkByItemId(article.id);

                  return (
                    <ArticleItem
                      key={article.id}
                      article={article}
                      selectedArticleId={selectedArticleId}
                      onSelectArticle={setSelectedArticle}
                      onToggleRead={handleToggleRead}
                      onToggleStar={handleToggleStar}
                      canToggleRead={article.id > 0}
                      isStarred={isItemStarred(article.id)}
                      feedName={feed?.name ?? bookmark?.feed_name ?? "Unknown"}
                      feedFaviconUrl={
                        feed ? getFaviconUrl(feed.link, feed.site_url) : null
                      }
                    />
                  );
                })}
                {hasMore && (
                  <div className="flex justify-center py-4">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => itemsQuery.fetchNextPage()}
                      disabled={isLoadingMore}
                      className="gap-2"
                    >
                      {isLoadingMore && (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      )}
                      {isLoadingMore ? "Loading..." : "Load more"}
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
