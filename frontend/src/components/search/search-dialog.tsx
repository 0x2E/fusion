import { useState, useEffect } from "react";
import { FileText, Search } from "lucide-react";
import { getFaviconUrl } from "@/lib/api/favicon";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command";
import { useUIStore, useDataStore } from "@/store";
import { useUrlState } from "@/hooks/use-url-state";
import { formatDate } from "@/lib/utils";

export function SearchDialog() {
  const { isSearchOpen, setSearchOpen } = useUIStore();
  const { setSelectedFeed, setSelectedArticle } = useUrlState();
  const { feeds, items, getFeedById } = useDataStore();
  const [query, setQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");

  // Debounce search query
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(query);
    }, 200);
    return () => clearTimeout(timer);
  }, [query]);

  // Reset query when dialog closes
  useEffect(() => {
    if (!isSearchOpen) {
      setQuery("");
      setDebouncedQuery("");
    }
  }, [isSearchOpen]);

  // Filter feeds by debounced query
  const filteredFeeds = debouncedQuery
    ? feeds.filter((feed) =>
        feed.name.toLowerCase().includes(debouncedQuery.toLowerCase()),
      )
    : [];

  // Filter articles by debounced query
  const filteredArticles = debouncedQuery
    ? items
        .filter((item) =>
          item.title.toLowerCase().includes(debouncedQuery.toLowerCase()),
        )
        .slice(0, 10)
    : [];

  const handleSelectFeed = (feedId: number) => {
    setSelectedFeed(feedId);
    setSearchOpen(false);
  };

  const handleSelectArticle = (articleId: number) => {
    setSelectedArticle(articleId);
    setSearchOpen(false);
  };

  return (
    <CommandDialog open={isSearchOpen} onOpenChange={setSearchOpen}>
      <CommandInput
        placeholder="Search feeds and articles..."
        value={query}
        onValueChange={setQuery}
      />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>

        {filteredFeeds.length > 0 && (
          <CommandGroup heading="Feeds">
            {filteredFeeds.map((feed) => (
              <CommandItem
                key={`feed-${feed.id}`}
                value={`feed-${feed.name}`}
                onSelect={() => handleSelectFeed(feed.id)}
                className="gap-2"
              >
                <img
                  src={getFaviconUrl(feed.link, feed.site_url)}
                  alt=""
                  className="h-4 w-4 shrink-0 rounded-sm"
                  loading="lazy"
                />
                <span>{feed.name}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        {filteredFeeds.length > 0 && filteredArticles.length > 0 && (
          <CommandSeparator />
        )}

        {filteredArticles.length > 0 && (
          <CommandGroup heading="Articles">
            {filteredArticles.map((article) => {
              const feed = getFeedById(article.feed_id);
              return (
                <CommandItem
                  key={`article-${article.id}`}
                  value={`article-${article.title}`}
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

        {!debouncedQuery && (
          <CommandGroup heading="Quick Actions">
            <CommandItem className="gap-2">
              <Search className="h-4 w-4 text-muted-foreground" />
              <span>Type to search feeds and articles</span>
            </CommandItem>
          </CommandGroup>
        )}
      </CommandList>
    </CommandDialog>
  );
}
