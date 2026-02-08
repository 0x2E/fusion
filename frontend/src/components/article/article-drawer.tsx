import {
  Circle,
  CircleCheck,
  ChevronLeft,
  ChevronRight,
  ExternalLink,
  Star,
  X,
} from "lucide-react";
import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useUrlState } from "@/hooks/use-url-state";
import { useDataStore } from "@/store";
import { useArticles } from "@/hooks/use-articles";
import { useStarred } from "@/hooks/use-starred";
import { useArticleNavigation } from "@/hooks/use-keyboard";
import { formatDate } from "@/lib/utils";
import { processArticleContent } from "@/lib/content";
import { getFaviconUrl } from "@/lib/api/favicon";

export function ArticleDrawer() {
  const { selectedArticleId, setSelectedArticle } = useUrlState();
  const { getItemById, getFeedById } = useDataStore();
  const { articles, markAsRead, markAsUnread } = useArticles();
  const { toggleStar, isStarred } = useStarred();

  const articleIds = articles.map((a) => a.id);
  const { goToNext, goToPrevious, hasNext, hasPrevious } =
    useArticleNavigation(articleIds);

  const article = selectedArticleId ? getItemById(selectedArticleId) : null;
  const feed = article ? getFeedById(article.feed_id) : null;
  const starred = article ? isStarred(article.id) : false;

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      setSelectedArticle(null);
    }
  };

  const handleToggleRead = async () => {
    if (!article) return;
    if (article.unread) {
      await markAsRead(article.id);
    } else {
      await markAsUnread(article.id);
    }
  };

  const handleToggleStar = async () => {
    if (!article) return;
    await toggleStar(article);
  };

  const handleOpenOriginal = () => {
    if (!article) return;
    window.open(article.link, "_blank", "noopener,noreferrer");
  };

  const getLinkDomain = (url: string) => {
    try {
      return new URL(url).hostname;
    } catch {
      return url;
    }
  };

  return (
    <Sheet open={selectedArticleId !== null} onOpenChange={handleOpenChange}>
      <SheetContent
        side="right"
        className="w-full max-w-[720px] p-0 sm:max-w-[720px]"
        showCloseButton={false}
      >
        {article && (
          <div className="flex h-full flex-col">
            {/* Header */}
            <div className="flex items-center justify-between border-b px-6 py-3">
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleToggleRead}
                  className="h-auto gap-1.5 px-2.5 py-1.5 text-[13px] font-medium text-muted-foreground"
                >
                  {article.unread ? (
                    <Circle className="h-4 w-4 text-muted-foreground" />
                  ) : (
                    <CircleCheck className="h-4 w-4 text-primary" />
                  )}
                  {article.unread ? "Mark as read" : "Mark as unread"}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleToggleStar}
                  className="h-auto gap-1.5 px-2.5 py-1.5 text-[13px] font-medium text-muted-foreground"
                >
                  <Star
                    className={`h-4 w-4 ${starred ? "fill-current text-amber-500" : ""}`}
                  />
                  Star
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleOpenOriginal}
                  className="h-auto gap-1.5 px-2.5 py-1.5 text-[13px] font-medium text-muted-foreground"
                >
                  <ExternalLink className="h-4 w-4" />
                  Original
                </Button>
              </div>

              <SheetTitle className="sr-only">{article.title}</SheetTitle>

              <Button
                variant="ghost"
                size="icon"
                onClick={() => setSelectedArticle(null)}
                className="h-[30px] w-[30px] rounded-md"
              >
                <X className="h-[18px] w-[18px] text-muted-foreground" />
              </Button>
            </div>

            {/* Content */}
            <ScrollArea className="min-h-0 flex-1">
              <article className="px-12 py-8">
                <div className="space-y-3">
                  <h1 className="text-[28px] font-bold leading-[1.3]">
                    {article.title}
                  </h1>
                  <div className="flex items-center gap-2 text-sm">
                    <span className="flex items-center gap-1.5 rounded bg-muted px-2 py-1 text-xs font-medium text-muted-foreground">
                      {feed && (
                        <img
                          src={getFaviconUrl(feed.link, feed.site_url)}
                          alt=""
                          className="h-3.5 w-3.5 shrink-0 rounded-sm"
                          loading="lazy"
                        />
                      )}
                      {feed?.name ?? "Unknown"}
                    </span>
                    <span className="text-muted-foreground/70">·</span>
                    <span className="text-muted-foreground">
                      {formatDate(article.pub_date)}
                    </span>
                    <span className="text-muted-foreground/70">·</span>
                    <a
                      href={article.link}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary hover:underline"
                    >
                      {getLinkDomain(article.link)}
                    </a>
                  </div>
                </div>

                <div
                  className="prose prose-neutral mt-6 max-w-none break-words dark:prose-invert"
                  dangerouslySetInnerHTML={{
                    __html: processArticleContent(
                      article.content,
                      article.link,
                    ),
                  }}
                />
              </article>
            </ScrollArea>

            {/* Footer - Navigation */}
            <div className="flex items-center justify-between border-t px-6 py-3">
              <Button
                variant="outline"
                size="sm"
                onClick={goToPrevious}
                disabled={!hasPrevious()}
                className="h-auto gap-1.5 px-3 py-2 text-[13px] font-medium text-muted-foreground"
              >
                <ChevronLeft className="h-4 w-4" />
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={goToNext}
                disabled={!hasNext()}
                className="h-auto gap-1.5 px-3 py-2 text-[13px] font-medium text-muted-foreground"
              >
                Next
                <ChevronRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
}
