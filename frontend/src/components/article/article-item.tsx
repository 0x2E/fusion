import { cn, formatDate, extractSummary } from "@/lib/utils";
import { useUIStore, useDataStore } from "@/store";
import type { Item } from "@/lib/api";

interface ArticleItemProps {
  article: Item;
}

export function ArticleItem({ article }: ArticleItemProps) {
  const { selectedArticleId, setSelectedArticle } = useUIStore();
  const { getFeedById } = useDataStore();
  const isSelected = selectedArticleId === article.id;
  const feed = getFeedById(article.feed_id);

  return (
    <button
      onClick={() => setSelectedArticle(article.id)}
      className={cn(
        "flex w-full flex-col gap-1 rounded-md border p-3 text-left transition-colors",
        isSelected
          ? "border-primary bg-accent"
          : "border-transparent hover:bg-accent/50",
        article.unread && "border-l-2 border-l-primary"
      )}
    >
      <h3
        className={cn(
          "line-clamp-2 text-sm",
          article.unread ? "font-medium" : "text-muted-foreground"
        )}
      >
        {article.title}
      </h3>
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <span className="truncate">{feed?.name ?? "Unknown"}</span>
        <span>·</span>
        <span className="shrink-0">{formatDate(article.pub_date)}</span>
      </div>
      <p className="line-clamp-2 text-xs text-muted-foreground">
        {extractSummary(article.content, 100)}
      </p>
    </button>
  );
}
