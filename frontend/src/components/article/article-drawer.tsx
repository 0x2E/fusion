import {
  BookmarkPlus,
  BookmarkCheck,
  Check,
  ChevronLeft,
  ChevronRight,
  ExternalLink,
  X,
  Circle,
} from "lucide-react";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { useUIStore, useDataStore } from "@/store";
import { useArticles } from "@/hooks/use-articles";
import { useStarred } from "@/hooks/use-starred";
import { useArticleNavigation } from "@/hooks/use-keyboard";
import { formatDate, sanitizeHTML } from "@/lib/utils";

export function ArticleDrawer() {
  const { selectedArticleId, setSelectedArticle } = useUIStore();
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

  return (
    <Sheet open={selectedArticleId !== null} onOpenChange={handleOpenChange}>
      <SheetContent
        side="right"
        className="w-full max-w-[720px] p-0 sm:max-w-[720px]"
      >
        {article && (
          <div className="flex h-full flex-col">
            {/* Header */}
            <SheetHeader className="flex flex-row items-center justify-between border-b px-4 py-3">
              <div className="flex items-center gap-1">
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={handleToggleRead}
                        className="h-8 w-8"
                      >
                        {article.unread ? (
                          <Check className="h-4 w-4" />
                        ) : (
                          <Circle className="h-4 w-4" />
                        )}
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>
                      {article.unread ? "Mark as read" : "Mark as unread"}
                    </TooltipContent>
                  </Tooltip>

                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={handleToggleStar}
                        className="h-8 w-8"
                      >
                        {starred ? (
                          <BookmarkCheck className="h-4 w-4 text-primary" />
                        ) : (
                          <BookmarkPlus className="h-4 w-4" />
                        )}
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>
                      {starred ? "Remove star" : "Star article"}
                    </TooltipContent>
                  </Tooltip>

                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={handleOpenOriginal}
                        className="h-8 w-8"
                      >
                        <ExternalLink className="h-4 w-4" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Open original</TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>

              <SheetTitle className="sr-only">{article.title}</SheetTitle>

              <Button
                variant="ghost"
                size="icon"
                onClick={() => setSelectedArticle(null)}
                className="h-8 w-8"
              >
                <X className="h-4 w-4" />
              </Button>
            </SheetHeader>

            {/* Content */}
            <ScrollArea className="flex-1">
              <article className="p-6">
                <h1 className="text-2xl font-bold leading-tight">
                  {article.title}
                </h1>
                <div className="mt-3 flex items-center gap-2 text-sm text-muted-foreground">
                  <span>{feed?.name ?? "Unknown"}</span>
                  <span>·</span>
                  <span>{formatDate(article.pub_date)}</span>
                </div>

                <Separator className="my-6" />

                <div
                  className="prose prose-sm max-w-none dark:prose-invert"
                  dangerouslySetInnerHTML={{
                    __html: sanitizeHTML(article.content),
                  }}
                />
              </article>
            </ScrollArea>

            {/* Footer - Navigation */}
            <div className="flex items-center justify-between border-t px-4 py-3">
              <Button
                variant="ghost"
                size="sm"
                onClick={goToPrevious}
                disabled={!hasPrevious()}
                className="gap-1"
              >
                <ChevronLeft className="h-4 w-4" />
                Previous
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={goToNext}
                disabled={!hasNext()}
                className="gap-1"
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
