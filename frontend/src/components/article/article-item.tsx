import { Circle, CircleCheck, Star, ExternalLink } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useI18n } from "@/lib/i18n";
import { cn, formatDate, extractSummary } from "@/lib/utils";
import type { Item } from "@/lib/api";
import { FeedFavicon } from "@/components/feed/feed-favicon";
import { toSafeExternalUrl } from "@/lib/safe-url";

interface ArticleItemProps {
  article: Item;
  selectedArticleId: number | null;
  onSelectArticle: (articleId: number | null) => void;
  onToggleRead: (article: Item) => Promise<void>;
  onToggleStar: (article: Item) => Promise<void>;
  canToggleRead: boolean;
  isStarred: boolean;
  feedName: string;
  feedFaviconUrl: string | null;
}

export function ArticleItem({
  article,
  selectedArticleId,
  onSelectArticle,
  onToggleRead,
  onToggleStar,
  canToggleRead,
  isStarred,
  feedName,
  feedFaviconUrl,
}: ArticleItemProps) {
  const { t } = useI18n();

  const isSelected = selectedArticleId === article.id;
  const safeArticleLink = toSafeExternalUrl(article.link);

  const handleToggleRead = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (!canToggleRead) return;

    try {
      await onToggleRead(article);
    } catch (error) {
      console.error("Failed to toggle read status:", error);
    }
  };

  const handleToggleStar = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await onToggleStar(article);
    } catch (error) {
      console.error("Failed to toggle star:", error);
    }
  };

  const handleOpenExternal = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (safeArticleLink) {
      window.open(safeArticleLink, "_blank", "noopener,noreferrer");
    }
  };

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={() => onSelectArticle(article.id)}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onSelectArticle(article.id);
        }
      }}
      className={cn(
        "group relative flex w-full cursor-pointer items-start gap-4 border-b px-4 py-4 text-left transition-colors hover:bg-accent/50",
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
          <FeedFavicon src={feedFaviconUrl} className="h-3.5 w-3.5 rounded-sm" />
          <span className="truncate font-medium text-muted-foreground">
            {feedName}
          </span>
          <span className="text-muted-foreground">·</span>
          <span className="shrink-0 text-muted-foreground">
            {formatDate(article.pub_date)}
          </span>
        </div>
      </div>

      {/* Article Actions */}
      <div className="absolute right-2 top-2 hidden items-center gap-1 group-hover:flex">
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={handleToggleRead}
          disabled={!canToggleRead}
          className={cn(article.unread ? "bg-muted" : "bg-primary/10")}
          title={
            article.unread
              ? t("article.action.markRead")
              : t("article.action.markUnread")
          }
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
          className={cn(isStarred ? "bg-amber-100 dark:bg-amber-950/40" : "bg-muted")}
          title={isStarred ? t("article.action.unstar") : t("article.action.star")}
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
          disabled={!safeArticleLink}
          className="bg-muted"
          title={t("article.action.openInBrowser")}
        >
          <ExternalLink className="text-muted-foreground" />
        </Button>
      </div>
    </div>
  );
}
