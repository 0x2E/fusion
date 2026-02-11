import { useMatchRoute } from "@tanstack/react-router";
import { Inbox, Layers, Star } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useGroups } from "@/queries/groups";
import { useFeedLookup, useUnreadCounts } from "@/queries/feeds";
import { useBookmarkLookup } from "@/queries/bookmarks";
import { useUrlState } from "@/hooks/use-url-state";
import { cn } from "@/lib/utils";
import { FeedGroup } from "./feed-group";
import { FeedItem } from "./feed-item";

export function FeedList() {
  const { data: groups = [], isLoading } = useGroups();
  const { feeds, getFeedsByGroup } = useFeedLookup();
  const { getTotalUnreadCount } = useUnreadCounts();
  const { bookmarks } = useBookmarkLookup();
  const {
    selectedFeedId,
    selectedGroupId,
    articleFilter,
    selectTopLevelFilter,
  } = useUrlState();
  const matchRoute = useMatchRoute();
  const isOnHomePage = !!matchRoute({ to: "/" });
  const isTopLevelSelected =
    isOnHomePage && selectedFeedId === null && selectedGroupId === null;
  const totalUnread = getTotalUnreadCount();
  const starredCount = bookmarks.length;

  const topFilters: Array<{
    value: "all" | "unread" | "starred";
    label: string;
    count: number;
    icon: typeof Inbox;
  }> = [
    { value: "unread", label: "Unread", count: totalUnread, icon: Inbox },
    { value: "starred", label: "Starred", count: starredCount, icon: Star },
    { value: "all", label: "All", count: totalUnread, icon: Layers },
  ];

  if (isLoading && groups.length === 0) {
    return (
      <div className="flex-1 p-4">
        <div className="space-y-2">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-8 animate-pulse rounded-md bg-accent" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <ScrollArea className="flex-1 w-full min-w-0 overflow-hidden [&_[data-slot=scroll-area-viewport]>div]:!block">
      <div className="w-full min-w-0 p-2 space-y-0.5">
        {/* Top-level filters */}
        <div className="space-y-0.5">
          {topFilters.map(({ value, label, count, icon: Icon }) => (
            <button
              key={value}
              onClick={() => selectTopLevelFilter(value)}
              className={cn(
                "flex w-full min-w-0 items-center gap-1.5 rounded-md px-2 py-1 text-left text-sm transition-colors",
                isTopLevelSelected && articleFilter === value
                  ? "bg-accent text-accent-foreground"
                  : "hover:bg-accent/50",
              )}
            >
              <Icon className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
              <span className="min-w-0 flex-1">{label}</span>
              <span className="shrink-0 text-[11px] text-muted-foreground">
                {count}
              </span>
            </button>
          ))}
        </div>

        {/* Feeds header */}
        <div className="mt-2 flex items-center justify-between px-2 py-1">
          <span className="text-[11px] font-medium text-muted-foreground">
            Feeds
          </span>
        </div>

        {/* Feed groups */}
        <div className="w-full min-w-0 space-y-0.5">
          {groups.map((group) => {
            const groupFeeds = getFeedsByGroup(group.id);

            return (
              <FeedGroup
                key={group.id}
                groupId={group.id}
                name={group.name}
                feeds={groupFeeds}
              />
            );
          })}

          {/* Ungrouped feeds (group_id = 0) */}
          {feeds
            .filter((f) => f.group_id === 0)
            .map((feed) => (
              <FeedItem key={feed.id} feed={feed} />
            ))}
        </div>
      </div>
    </ScrollArea>
  );
}
