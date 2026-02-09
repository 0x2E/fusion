import { useMatchRoute } from "@tanstack/react-router";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useGroups } from "@/queries/groups";
import { useFeedLookup, useUnreadCounts } from "@/queries/feeds";
import { useUrlState } from "@/hooks/use-url-state";
import { cn } from "@/lib/utils";
import { Inbox } from "lucide-react";
import { FeedGroup } from "./feed-group";
import { FeedItem } from "./feed-item";

export function FeedList() {
  const { data: groups = [], isLoading } = useGroups();
  const { feeds, getFeedsByGroup } = useFeedLookup();
  const { getTotalUnreadCount } = useUnreadCounts();
  const { selectedFeedId, selectedGroupId, selectAll } = useUrlState();
  const matchRoute = useMatchRoute();
  const isOnHomePage = !!matchRoute({ to: "/" });
  const isAllSelected =
    isOnHomePage && selectedFeedId === null && selectedGroupId === null;
  const totalUnread = getTotalUnreadCount();

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
        {/* Feeds header */}
        <div className="flex items-center justify-between px-2 py-1">
          <span className="text-xs font-medium text-muted-foreground">
            Feeds
          </span>
        </div>

        {/* All feeds */}
        <button
          onClick={selectAll}
          className={cn(
            "flex w-full min-w-0 items-center gap-1.5 rounded-md px-2 py-1 text-left text-sm transition-colors",
            isAllSelected
              ? "bg-accent text-accent-foreground"
              : "hover:bg-accent/50",
          )}
        >
          <Inbox className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
          <span className="min-w-0 flex-1">All</span>
          <span className="shrink-0 text-xs text-muted-foreground">
            {totalUnread}
          </span>
        </button>

        {/* Feed groups */}
        <div className="mt-1 w-full min-w-0 space-y-0.5">
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
