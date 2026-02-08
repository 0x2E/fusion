import { Circle, CircleCheck, Star, ExternalLink } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn, formatDate, extractSummary } from "@/lib/utils";
import { useUrlState } from "@/hooks/use-url-state";
import { useDataStore } from "@/store";
import { itemAPI, bookmarkAPI, type Item } from "@/lib/api";
import { getFaviconUrl } from "@/lib/api/favicon";

interface ArticleItemProps {
  article: Item;
}

export function ArticleItem({ article }: ArticleItemProps) {
  const { selectedArticleId, setSelectedArticle } = useUrlState();
  const {
    getFeedById,
    markItemRead,
    markItemUnread,
    isItemStarred,
    getBookmarkByItemId,
    addBookmark,
    removeBookmark,
  } = useDataStore();
  const isSelected = selectedArticleId === article.id;
  const feed = getFeedById(article.feed_id);
  const isStarred = isItemStarred(article.id);

  const handleToggleRead = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      if (article.unread) {
        await itemAPI.markRead({ ids: [article.id] });
        markItemRead(article.id);
      } else {
        await itemAPI.markUnread({ ids: [article.id] });
        markItemUnread(article.id);
      }
    } catch (error) {
      console.error("Failed to toggle read status:", error);
    }
  };

  const handleToggleStar = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      if (isStarred) {
        const bookmark = getBookmarkByItemId(article.id);
        if (bookmark) {
          await bookmarkAPI.delete(bookmark.id);
          removeBookmark(bookmark.id);
        }
      } else {
        const response = await bookmarkAPI.create({
          item_id: article.id,
          link: article.link,
          title: article.title,
          content: article.content,
          pub_date: article.pub_date,
          feed_name: feed?.name ?? "Unknown",
        });
        if (response.data) {
          addBookmark(response.data);
        }
      }
    } catch (error) {
      console.error("Failed to toggle star:", error);
    }
  };

  const handleOpenExternal = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (article.link) {
      window.open(article.link, "_blank", "noopener,noreferrer");
    }
  };

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={() => setSelectedArticle(article.id)}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          setSelectedArticle(article.id);
        }
      }}
      className={cn(
        "group flex w-full cursor-pointer items-start gap-4 border-b px-4 py-4 text-left transition-colors hover:bg-accent/50",
        isSelected && "bg-accent",
      )}
    >
      {/* Article Content */}
      <div className="flex min-w-0 flex-1 flex-col gap-1.5">
        <h3
          className={cn(
            "line-clamp-2 text-[15px] leading-snug font-medium",
            article.unread ? "text-foreground" : "text-muted-foreground",
          )}
        >
          {article.title}
        </h3>
        <p className="line-clamp-2 text-sm text-muted-foreground">
          {extractSummary(article.content, 150)}
        </p>
        <div className="flex items-center gap-2 text-xs">
          {feed && (
            <img
              src={getFaviconUrl(feed.link, feed.site_url)}
              alt=""
              className="h-3.5 w-3.5 shrink-0 rounded-sm"
              loading="lazy"
            />
          )}
          <span className="truncate font-medium text-muted-foreground">
            {feed?.name ?? "Unknown"}
          </span>
          <span className="text-muted-foreground">·</span>
          <span className="shrink-0 text-muted-foreground">
            {formatDate(article.pub_date)}
          </span>
        </div>
      </div>

      {/* Article Actions */}
      <div className="flex shrink-0 items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={handleToggleRead}
          className={cn(article.unread ? "bg-muted" : "bg-primary/10")}
          title={article.unread ? "Mark as read" : "Mark as unread"}
        >
          {article.unread ? (
            <Circle className="text-muted-foreground" />
          ) : (
            <CircleCheck className="text-primary" />
          )}
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={handleToggleStar}
          className={cn(isStarred ? "bg-amber-50" : "bg-muted")}
          title={isStarred ? "Unstar" : "Star"}
        >
          <Star
            className={cn(
              isStarred
                ? "fill-amber-500 text-amber-500"
                : "text-muted-foreground",
            )}
          />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={handleOpenExternal}
          className="bg-muted"
          title="Open in browser"
        >
          <ExternalLink className="text-muted-foreground" />
        </Button>
      </div>
    </div>
  );
}
