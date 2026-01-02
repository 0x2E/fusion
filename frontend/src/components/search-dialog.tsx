import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useDataStore } from "@/store/data";
import { useNavigate } from "@tanstack/react-router";
import { FileText, Rss, Search } from "lucide-react";
import { useDeferredValue, useMemo, useState } from "react";

interface SearchDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function SearchDialog({ open, onOpenChange }: SearchDialogProps) {
  const navigate = useNavigate();
  const { feeds, items } = useDataStore();
  const [searchQuery, setSearchQuery] = useState("");
  const [searchType, setSearchType] = useState<"all" | "feeds" | "articles">(
    "all"
  );
  const deferredQuery = useDeferredValue(searchQuery);

  const searchResults = useMemo(() => {
    const query = deferredQuery.toLowerCase().trim();
    if (!query) {
      return { feeds: [], articles: [] };
    }

    const matchedFeeds = feeds.filter((feed) => {
      return (
        feed.name.toLowerCase().includes(query) ||
        feed.link.toLowerCase().includes(query) ||
        feed.site_url?.toLowerCase().includes(query)
      );
    });

    const matchedArticles = items.filter((item) => {
      return (
        item.title.toLowerCase().includes(query) ||
        item.content.toLowerCase().includes(query)
      );
    });

    return {
      feeds: matchedFeeds.slice(0, 20),
      articles: matchedArticles.slice(0, 20),
    };
  }, [deferredQuery, feeds, items]);

  const handleSelectFeed = (feedId: number) => {
    navigate({ to: "/", search: { feed: feedId, filter: "all" as const } });
    onOpenChange(false);
  };

  const handleSelectArticle = (itemId: number) => {
    navigate({ to: "/", search: { item: itemId } });
    onOpenChange(false);
  };

  const displayResults = useMemo(() => {
    if (searchType === "feeds") {
      return { feeds: searchResults.feeds, articles: [] };
    } else if (searchType === "articles") {
      return { feeds: [], articles: searchResults.articles };
    }
    return searchResults;
  }, [searchType, searchResults]);

  const totalResults =
    displayResults.feeds.length + displayResults.articles.length;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl p-0 gap-0">
        <DialogHeader className="px-6 pt-6 pb-4">
          <DialogTitle>Search</DialogTitle>
        </DialogHeader>

        <div className="px-6 pb-4 space-y-3">
          {/* Search Input */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              placeholder="Search feeds and articles..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9"
              autoFocus
            />
          </div>

          {/* Search Type Filters */}
          <div className="flex gap-1">
            <Button
              size="sm"
              variant={searchType === "all" ? "secondary" : "ghost"}
              onClick={() => setSearchType("all")}
            >
              All
            </Button>
            <Button
              size="sm"
              variant={searchType === "feeds" ? "secondary" : "ghost"}
              onClick={() => setSearchType("feeds")}
            >
              <Rss className="w-3 h-3 mr-1" />
              Feeds
            </Button>
            <Button
              size="sm"
              variant={searchType === "articles" ? "secondary" : "ghost"}
              onClick={() => setSearchType("articles")}
            >
              <FileText className="w-3 h-3 mr-1" />
              Articles
            </Button>
          </div>
        </div>

        {/* Search Results */}
        <ScrollArea className="max-h-96 border-t border-border">
          <div className="p-4">
            {!deferredQuery ? (
              <div className="text-center py-12 text-muted-foreground">
                <Search className="w-12 h-12 mx-auto mb-3 opacity-20" />
                <p>Start typing to search...</p>
              </div>
            ) : totalResults === 0 ? (
              <div className="text-center py-12 text-muted-foreground">
                <Search className="w-12 h-12 mx-auto mb-3 opacity-20" />
                <p>No results found for "{deferredQuery}"</p>
              </div>
            ) : (
              <div className="space-y-4">
                {/* Feeds Results */}
                {displayResults.feeds.length > 0 && (
                  <div className="space-y-1">
                    <div className="px-2 py-1 text-xs font-medium text-muted-foreground uppercase tracking-wider">
                      Feeds ({displayResults.feeds.length})
                    </div>
                    <div className="space-y-0.5">
                      {displayResults.feeds.map((feed) => (
                        <button
                          key={feed.id}
                          onClick={() => handleSelectFeed(feed.id)}
                          className="w-full text-left px-3 py-2 rounded-md hover:bg-accent transition-colors"
                        >
                          <div className="flex items-start gap-2">
                            <Rss className="w-4 h-4 mt-0.5 text-muted-foreground shrink-0" />
                            <div className="flex-1 min-w-0">
                              <div className="font-medium text-sm truncate">
                                {feed.name}
                              </div>
                              <div className="text-xs text-muted-foreground truncate">
                                {feed.link}
                              </div>
                            </div>
                          </div>
                        </button>
                      ))}
                    </div>
                  </div>
                )}

                {/* Articles Results */}
                {displayResults.articles.length > 0 && (
                  <div className="space-y-1">
                    <div className="px-2 py-1 text-xs font-medium text-muted-foreground uppercase tracking-wider">
                      Articles ({displayResults.articles.length})
                    </div>
                    <div className="space-y-0.5">
                      {displayResults.articles.map((article) => (
                        <button
                          key={article.id}
                          onClick={() => handleSelectArticle(article.id)}
                          className="w-full text-left px-3 py-2 rounded-md hover:bg-accent transition-colors"
                        >
                          <div className="flex items-start gap-2">
                            <FileText className="w-4 h-4 mt-0.5 text-muted-foreground shrink-0" />
                            <div className="flex-1 min-w-0">
                              <div className="font-medium text-sm line-clamp-2">
                                {article.title}
                              </div>
                              <div className="text-xs text-muted-foreground line-clamp-1 mt-0.5">
                                {new Date(
                                  article.pub_date * 1000
                                ).toLocaleDateString()}
                              </div>
                            </div>
                          </div>
                        </button>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
}
