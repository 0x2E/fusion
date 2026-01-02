import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from "@/components/ui/empty";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import type { Item } from "@/lib/api/types";
import { cn } from "@/lib/utils";
import { useDataStore } from "@/store/data";
import { useNavigate, useSearch } from "@tanstack/react-router";
import {
  ArrowUpDown,
  Check,
  CheckCheck,
  Circle,
  ExternalLink,
  Menu,
  MoreHorizontal,
  Pause,
  Settings,
  Star,
  Trash2,
} from "lucide-react";
import { useEffect } from "react";

interface ArticleListProps {
  onMenuOpen?: () => void;
}

export function ArticleList({ onMenuOpen }: ArticleListProps) {
  const navigate = useNavigate();
  const search = useSearch({ from: "/" });
  const {
    items,
    feeds,
    groups,
    itemsLoading,
    fetchItems,
    markItemsRead,
    markItemsUnread,
  } = useDataStore();

  const filter = search.filter || "all";

  useEffect(() => {
    const params = {
      feed_id: search.feed,
      group_id: search.group,
      unread: filter === "unread" ? true : undefined,
      limit: 100,
    };
    fetchItems(params);
  }, [search.feed, search.group, filter, fetchItems]);

  const handleFilterChange = (newFilter: "all" | "unread" | "starred") => {
    navigate({
      to: "/",
      search: {
        feed: search.feed,
        group: search.group,
        filter: newFilter,
        search: search.search,
        settings: search.settings,
      },
    });
  };

  const handleSelectArticle = (itemId: number) => {
    navigate({
      to: "/",
      search: {
        feed: search.feed,
        group: search.group,
        filter: search.filter,
        item: itemId,
        search: search.search,
        settings: search.settings,
      },
    });
  };

  const handleToggleRead = async (item: Item, e: React.MouseEvent) => {
    e.stopPropagation();
    if (item.unread) {
      await markItemsRead([item.id]);
    } else {
      await markItemsUnread([item.id]);
    }
  };

  const handleMarkAllRead = async () => {
    const unreadIds = items
      .filter((item) => item.unread)
      .map((item) => item.id);
    if (unreadIds.length > 0) {
      await markItemsRead(unreadIds);
    }
  };

  const handleOpenLink = (link: string, e: React.MouseEvent) => {
    e.stopPropagation();
    window.open(link, "_blank");
  };

  const getFeedName = (feedId: number) => {
    return feeds.find((f) => f.id === feedId)?.name || "Unknown";
  };

  const formatDate = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleDateString();
  };

  const getCurrentViewName = () => {
    if (search.feed) {
      return feeds.find((f) => f.id === search.feed)?.name || "Feed";
    }
    if (search.group) {
      return groups.find((g) => g.id === search.group)?.name || "Group";
    }
    return "All Articles";
  };

  const feedType = search.feed ? "feed" : search.group ? "group" : "all";

  const filteredItems = filter === "starred" ? [] : items;

  return (
    <div className="h-full flex flex-col">
      <div className="border-b border-border px-6 py-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-3">
            {onMenuOpen && (
              <Button
                variant="ghost"
                size="sm"
                className="md:hidden -ml-2"
                onClick={onMenuOpen}
                title="Open menu"
              >
                <Menu className="w-4 h-4" />
              </Button>
            )}
            <h2 className="text-lg font-semibold text-foreground">
              {getCurrentViewName()}
            </h2>
          </div>

          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              title="Mark all as read"
              onClick={handleMarkAllRead}
            >
              <CheckCheck className="w-4 h-4" />
            </Button>

            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" title="Sort">
                  <ArrowUpDown className="w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem>Newest first</DropdownMenuItem>
                <DropdownMenuItem>Oldest first</DropdownMenuItem>
                <DropdownMenuItem>By feed name</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>

            {feedType === "group" && (
              <Button
                variant="ghost"
                size="sm"
                title="Pause group"
                onClick={() => {}}
              >
                <Pause className="w-4 h-4" />
              </Button>
            )}

            {feedType === "feed" && (
              <>
                <Button
                  variant="ghost"
                  size="sm"
                  title="Pause feed"
                  onClick={() => {}}
                >
                  <Pause className="w-4 h-4" />
                </Button>

                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="sm" title="More options">
                      <MoreHorizontal className="w-4 h-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem>
                      <Settings className="w-4 h-4 mr-2" />
                      Settings
                    </DropdownMenuItem>
                    <DropdownMenuItem className="text-destructive">
                      <Trash2 className="w-4 h-4 mr-2" />
                      Delete Feed
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </>
            )}
          </div>
        </div>

        <div className="flex gap-1">
          <Button
            variant={filter === "all" ? "secondary" : "ghost"}
            size="sm"
            onClick={() => handleFilterChange("all")}
          >
            All
          </Button>
          <Button
            variant={filter === "unread" ? "secondary" : "ghost"}
            size="sm"
            onClick={() => handleFilterChange("unread")}
          >
            Unread
          </Button>
          <Button
            variant={filter === "starred" ? "secondary" : "ghost"}
            size="sm"
            onClick={() => handleFilterChange("starred")}
          >
            Starred
          </Button>
        </div>
      </div>

      <ScrollArea className="flex-1">
        {itemsLoading ? (
          <div className="space-y-2 p-4">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
          </div>
        ) : filteredItems.length === 0 ? (
          <Empty>
            <EmptyHeader>
              <EmptyTitle>No articles</EmptyTitle>
              <EmptyDescription>
                There are no articles to display.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <div className="divide-y divide-border">
            {filteredItems.map((item) => {
              const isBookmarked = false;

              return (
                <div
                  key={item.id}
                  className="group/article relative cursor-pointer transition-colors hover:bg-muted/30"
                  onClick={() => handleSelectArticle(item.id)}
                  onContextMenu={(e) => {
                    e.preventDefault();
                    const button = e.currentTarget.querySelector(
                      "[data-context-trigger]"
                    ) as HTMLElement;
                    button?.click();
                  }}
                >
                  <div className="px-6 py-3 flex flex-col md:grid md:grid-cols-[20px_auto_130px_100px] md:gap-4 md:items-center">
                    <div className="flex items-start gap-3 md:contents">
                      <div className="flex items-center justify-start w-5 shrink-0 pt-0.5 md:pt-0">
                        {item.unread && isBookmarked && (
                          <div className="relative w-3.5 h-3.5">
                            <Circle className="absolute w-2 h-2 text-primary fill-primary top-0 left-0" />
                            <Star className="absolute w-2.5 h-2.5 fill-yellow-500 text-yellow-500 bottom-0 right-0" />
                          </div>
                        )}
                        {item.unread && !isBookmarked && (
                          <Circle className="w-2 h-2 text-primary fill-primary" />
                        )}
                        {!item.unread && isBookmarked && (
                          <Star className="w-3 h-3 fill-yellow-500 text-yellow-500" />
                        )}
                      </div>

                      <div className="flex-1 min-w-0 md:pr-4">
                        <h3
                          className={cn(
                            "font-medium text-sm leading-snug mb-1 line-clamp-2 md:line-clamp-1",
                            !item.unread
                              ? "text-muted-foreground"
                              : "text-foreground"
                          )}
                        >
                          {item.title}
                        </h3>
                        <p className="hidden md:block text-xs text-muted-foreground truncate line-clamp-1">
                          {item.content
                            .replace(/<[^>]*>/g, "")
                            .substring(0, 150)}
                        </p>
                      </div>

                      <span className="hidden md:block text-xs text-muted-foreground truncate text-right">
                        {getFeedName(item.feed_id)}
                      </span>
                      <span className="hidden md:block text-xs text-muted-foreground text-right">
                        {formatDate(item.pub_date)}
                      </span>
                    </div>

                    <div className="flex items-center gap-2 mt-1 ml-8 text-xs text-muted-foreground md:hidden">
                      <span className="truncate">
                        {getFeedName(item.feed_id)}
                      </span>
                      <span className="text-muted-foreground/60">•</span>
                      <span className="shrink-0">
                        {formatDate(item.pub_date)}
                      </span>
                    </div>

                    <div className="absolute right-0 top-0 bottom-0 opacity-0 group-hover/article:opacity-100 transition-opacity flex items-stretch">
                      <div className="flex items-center gap-1 bg-background/95 backdrop-blur-sm px-2 h-full">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <button data-context-trigger className="hidden" />
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem
                              onClick={(e) => handleToggleRead(item, e)}
                            >
                              <Check className="w-3.5 h-3.5 mr-2" />
                              {!item.unread ? "Mark as Unread" : "Mark as Read"}
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={(e) => handleOpenLink(item.link, e)}
                            >
                              <ExternalLink className="w-3.5 h-3.5 mr-2" />
                              Open Link
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>

                        <Button
                          size="sm"
                          variant="ghost"
                          className="h-7 w-7 p-0 hover:bg-muted"
                          title={
                            !item.unread ? "Mark as Unread" : "Mark as Read"
                          }
                          onClick={(e) => handleToggleRead(item, e)}
                        >
                          {!item.unread ? (
                            <Circle className="w-3.5 h-3.5" />
                          ) : (
                            <Check className="w-3.5 h-3.5" />
                          )}
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          className="h-7 w-7 p-0 hover:bg-muted"
                          title="Open Link"
                          onClick={(e) => handleOpenLink(item.link, e)}
                        >
                          <ExternalLink className="w-3.5 h-3.5" />
                        </Button>
                      </div>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </ScrollArea>
    </div>
  );
}
