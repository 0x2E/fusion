import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import { useDataStore } from "@/store/data";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { Check, Circle, ExternalLink, Star } from "lucide-react";
import { useEffect, useState } from "react";

export function ArticleDrawer() {
  const navigate = useNavigate();
  const search = useSearch({ from: "/" });
  const {
    items,
    feeds,
    bookmarks,
    markItemsRead,
    markItemsUnread,
    createBookmark,
    deleteBookmark,
  } = useDataStore();
  const [isCreatingBookmark, setIsCreatingBookmark] = useState(false);

  const itemId = search.item;
  const open = !!itemId;
  const article = items.find((item) => item.id === itemId);
  const feed = article
    ? feeds.find((f) => f.id === article.feed_id)
    : undefined;
  const bookmark = article
    ? bookmarks.find((b) => b.item_id === article.id)
    : undefined;

  useEffect(() => {
    if (article && article.unread) {
      markItemsRead([article.id]);
    }
  }, [article, markItemsRead]);

  const handleClose = () => {
    navigate({
      to: "/",
      search: {
        feed: search.feed,
        group: search.group,
        filter: search.filter,
        search: search.search,
        settings: search.settings,
      },
    });
  };

  const handleToggleRead = async () => {
    if (!article) return;
    if (article.unread) {
      await markItemsRead([article.id]);
    } else {
      await markItemsUnread([article.id]);
    }
  };

  const handleToggleBookmark = async () => {
    if (!article) return;
    setIsCreatingBookmark(true);
    try {
      if (bookmark) {
        await deleteBookmark(bookmark.id);
      } else {
        await createBookmark({
          item_id: article.id,
          link: article.link,
          title: article.title,
          content: article.content,
          pub_date: article.pub_date,
          feed_name: feed?.name || "Unknown",
        });
      }
    } finally {
      setIsCreatingBookmark(false);
    }
  };

  const handleOpenLink = () => {
    if (article) {
      window.open(article.link, "_blank");
    }
  };

  const formatDate = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  if (!itemId) return null;

  return (
    <Sheet open={open} onOpenChange={handleClose}>
      <SheetContent className="w-full sm:max-w-3xl p-0 flex flex-col">
        {!article ? (
          <div className="flex-1 flex items-center justify-center">
            <div className="space-y-4 w-full max-w-2xl px-8">
              <Skeleton className="h-8 w-3/4" />
              <Skeleton className="h-4 w-1/2" />
              <Skeleton className="h-64 w-full" />
            </div>
          </div>
        ) : (
          <>
            <SheetHeader className="border-b border-border px-8 py-6 space-y-4">
              <div className="flex items-center gap-1">
                <Button
                  size="icon"
                  variant="ghost"
                  className="h-8 w-8"
                  title={!article.unread ? "Mark as Unread" : "Mark as Read"}
                  onClick={handleToggleRead}
                >
                  {!article.unread ? (
                    <Circle className="w-4 h-4" />
                  ) : (
                    <Check className="w-4 h-4" />
                  )}
                </Button>
                <Button
                  size="icon"
                  variant="ghost"
                  className="h-8 w-8"
                  title={bookmark ? "Remove Bookmark" : "Add Bookmark"}
                  onClick={handleToggleBookmark}
                  disabled={isCreatingBookmark}
                >
                  <Star
                    className={`w-4 h-4 ${
                      bookmark && "fill-yellow-500 text-yellow-500"
                    }`}
                  />
                </Button>
                <Button
                  size="icon"
                  variant="ghost"
                  className="h-8 w-8"
                  title="Open Original Link"
                  onClick={handleOpenLink}
                >
                  <ExternalLink className="w-4 h-4" />
                </Button>
              </div>

              <div className="space-y-2">
                <SheetTitle className="text-3xl font-bold text-balance leading-tight tracking-tight">
                  {article.title}
                </SheetTitle>
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <span>{feed?.name || "Unknown"}</span>
                  <span>•</span>
                  <span>{formatDate(article.pub_date)}</span>
                </div>
              </div>
            </SheetHeader>

            <div className="flex-1 relative">
              <ScrollArea className="h-full">
                <article className="max-w-3xl mx-auto px-16 py-12">
                  <div
                    className="notion-content space-y-4 text-foreground leading-relaxed"
                    dangerouslySetInnerHTML={{ __html: article.content }}
                  />
                </article>
              </ScrollArea>
            </div>
          </>
        )}
      </SheetContent>
    </Sheet>
  );
}
