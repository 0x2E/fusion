import { CheckCheck, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ArticleItem } from "./article-item";
import { SidebarTrigger } from "@/components/layout/sidebar-trigger";
import { useArticles } from "@/hooks/use-articles";
import { useArticleNavigation } from "@/hooks/use-keyboard";
import { useUrlState, type ArticleFilter } from "@/hooks/use-url-state";
import { useDataStore } from "@/store";

export function ArticleList() {
  const {
    articles,
    isLoading,
    isLoadingMore,
    hasMore,
    loadMore,
    markAllAsRead,
  } = useArticles();
  const { articleFilter, setArticleFilter, selectedFeedId, selectedGroupId } =
    useUrlState();
  const { items, getFeedById, getGroupById } = useDataStore();

  // Setup keyboard navigation
  const articleIds = articles.map((a) => a.id);
  useArticleNavigation(articleIds);

  // Determine title based on selection
  let title = "All Articles";
  if (selectedFeedId) {
    const feed = getFeedById(selectedFeedId);
    title = feed?.name ?? "Feed";
  } else if (selectedGroupId) {
    const group = getGroupById(selectedGroupId);
    title = group?.name ?? "Group";
  }

  const unreadCount = articles.filter((a) => a.unread).length;

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
          onClick={markAllAsRead}
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
        {items.length > 0 && (
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
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <p className="text-sm text-muted-foreground">
                  No articles found
                </p>
              </div>
            ) : (
              <>
                {articles.map((article) => (
                  <ArticleItem key={article.id} article={article} />
                ))}
                {hasMore && (
                  <div className="flex justify-center py-4">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={loadMore}
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
