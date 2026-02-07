import { useState, useEffect } from "react";
import { FileText, Loader2, Search, Settings } from "lucide-react";
import { getFaviconUrl } from "@/lib/api/favicon";
import { searchAPI } from "@/lib/api";
import type { SearchFeed, SearchItem } from "@/lib/api/types";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { useUIStore, useDataStore } from "@/store";
import { useUrlState } from "@/hooks/use-url-state";
import { formatDate } from "@/lib/utils";

export function SearchDialog() {
  const { isSearchOpen, setSearchOpen, setEditFeedOpen } = useUIStore();
  const { getFeedById } = useDataStore();
  const { setSelectedFeed, setSelectedArticle } = useUrlState();
  const [query, setQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [feeds, setFeeds] = useState<SearchFeed[]>([]);
  const [items, setItems] = useState<SearchItem[]>([]);

  // Debounce search query
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(query);
    }, 500);
    return () => clearTimeout(timer);
  }, [query]);

  // Reset state when dialog closes
  useEffect(() => {
    if (!isSearchOpen) {
      setQuery("");
      setDebouncedQuery("");
      setFeeds([]);
      setItems([]);
    }
  }, [isSearchOpen]);

  // Fetch search results from backend
  useEffect(() => {
    if (!debouncedQuery) {
      setFeeds([]);
      setItems([]);
      return;
    }

    let cancelled = false;
    setLoading(true);

    searchAPI
      .search(debouncedQuery)
      .then((res) => {
        if (cancelled) return;
        setFeeds(res.data?.feeds ?? []);
        setItems(res.data?.items ?? []);
      })
      .catch(() => {
        if (cancelled) return;
        setFeeds([]);
        setItems([]);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [debouncedQuery]);

  const handleSelectFeed = (feedId: number) => {
    setSelectedFeed(feedId);
    setSearchOpen(false);
  };

  const handleEditFeed = (e: React.MouseEvent, feedId: number) => {
    e.stopPropagation();
    const fullFeed = getFeedById(feedId);
    if (fullFeed) {
      setSearchOpen(false);
      setEditFeedOpen(true, fullFeed);
    }
  };

  const handleSelectArticle = (articleId: number) => {
    setSelectedArticle(articleId);
    setSearchOpen(false);
  };

  return (
    <Dialog open={isSearchOpen} onOpenChange={setSearchOpen}>
      <DialogHeader className="sr-only">
        <DialogTitle>Search</DialogTitle>
        <DialogDescription>Search feeds and articles</DialogDescription>
      </DialogHeader>
      <DialogContent className="overflow-hidden p-0" showCloseButton={false}>
        <Command
          shouldFilter={false}
          className="[&_[cmdk-group-heading]]:text-muted-foreground **:data-[slot=command-input-wrapper]:h-12 [&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:font-medium [&_[cmdk-group]]:px-2 [&_[cmdk-group]:not([hidden])_~[cmdk-group]]:pt-0 [&_[cmdk-input-wrapper]_svg]:h-5 [&_[cmdk-input-wrapper]_svg]:w-5 [&_[cmdk-input]]:h-12 [&_[cmdk-item]]:px-2 [&_[cmdk-item]]:py-3 [&_[cmdk-item]_svg]:h-5 [&_[cmdk-item]_svg]:w-5"
        >
          <CommandInput
            placeholder="Search feeds and articles..."
            value={query}
            onValueChange={setQuery}
          />
          <CommandList>
            {loading && debouncedQuery && (
              <div className="flex items-center justify-center py-6">
                <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
              </div>
            )}

            {!loading &&
              debouncedQuery &&
              feeds.length === 0 &&
              items.length === 0 && (
                <CommandEmpty>No results found.</CommandEmpty>
              )}

            {feeds.length > 0 && (
              <CommandGroup heading="Feeds">
                {feeds.map((feed) => (
                  <CommandItem
                    key={`feed-${feed.id}`}
                    onSelect={() => handleSelectFeed(feed.id)}
                    className="group gap-2"
                  >
                    <img
                      src={getFaviconUrl(feed.link, feed.site_url)}
                      alt=""
                      className="h-4 w-4 shrink-0 rounded-sm"
                      loading="lazy"
                    />
                    <span className="flex-1 truncate">{feed.name}</span>
                    <Button
                      variant="outline"
                      size="icon"
                      className="h-6 w-6"
                      onClick={(e) => handleEditFeed(e, feed.id)}
                    >
                      <Settings className="h-3.5 w-3.5 text-muted-foreground" />
                    </Button>
                  </CommandItem>
                ))}
              </CommandGroup>
            )}

            {feeds.length > 0 && items.length > 0 && <CommandSeparator />}

            {items.length > 0 && (
              <CommandGroup heading="Articles">
                {items.map((article) => {
                  const feed = feeds.find((f) => f.id === article.feed_id);
                  return (
                    <CommandItem
                      key={`article-${article.id}`}
                      onSelect={() => handleSelectArticle(article.id)}
                      className="flex-col items-start gap-1"
                    >
                      <div className="flex w-full items-center gap-2">
                        <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
                        <span className="flex-1 truncate">{article.title}</span>
                      </div>
                      <div className="flex w-full items-center gap-2 pl-6 text-xs text-muted-foreground">
                        <span>{feed?.name ?? "Unknown"}</span>
                        <span>·</span>
                        <span>{formatDate(article.pub_date)}</span>
                      </div>
                    </CommandItem>
                  );
                })}
              </CommandGroup>
            )}

            {!debouncedQuery && !loading && (
              <CommandGroup heading="Quick Actions">
                <CommandItem className="gap-2">
                  <Search className="h-4 w-4 text-muted-foreground" />
                  <span>Type to search feeds and articles</span>
                </CommandItem>
              </CommandGroup>
            )}
          </CommandList>
        </Command>
      </DialogContent>
    </Dialog>
  );
}
