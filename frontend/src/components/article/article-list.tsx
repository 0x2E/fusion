import { CheckCheck } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ArticleItem } from "./article-item";
import { useArticles } from "@/hooks/use-articles";
import { useArticleNavigation } from "@/hooks/use-keyboard";
import { useUIStore, useDataStore, type ArticleFilter } from "@/store";

export function ArticleList() {
  const { articles, isLoading, markAllAsRead } = useArticles();
  const { articleFilter, setArticleFilter, selectedFeedId, selectedGroupId } = useUIStore();
  const { getFeedById, getGroupById } = useDataStore();

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
      <div className="flex items-center justify-between border-b px-4 py-3">
        <h2 className="text-lg font-semibold">{title}</h2>
        <Button
          variant="ghost"
          size="sm"
          onClick={markAllAsRead}
          disabled={unreadCount === 0}
          className="gap-1.5 text-xs"
        >
          <CheckCheck className="h-4 w-4" />
          Mark all as read
        </Button>
      </div>

      {/* Filter tabs */}
      <div className="border-b px-4 py-2">
        <Tabs
          value={articleFilter}
          onValueChange={(v) => setArticleFilter(v as ArticleFilter)}
        >
          <TabsList className="h-8">
            <TabsTrigger value="all" className="text-xs">
              All
            </TabsTrigger>
            <TabsTrigger value="unread" className="text-xs">
              Unread
            </TabsTrigger>
            <TabsTrigger value="starred" className="text-xs">
              Starred
            </TabsTrigger>
          </TabsList>
        </Tabs>
      </div>

      {/* Article list */}
      <ScrollArea className="flex-1">
        <div className="space-y-1 p-2">
          {isLoading && articles.length === 0 ? (
            <div className="space-y-2 p-2">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="h-24 animate-pulse rounded-md bg-accent" />
              ))}
            </div>
          ) : articles.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <p className="text-sm text-muted-foreground">No articles found</p>
            </div>
          ) : (
            articles.map((article) => (
              <ArticleItem key={article.id} article={article} />
            ))
          )}
        </div>
      </ScrollArea>
    </div>
  );
}
