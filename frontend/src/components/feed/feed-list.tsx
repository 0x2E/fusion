import { Inbox } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { FeedGroup } from "./feed-group";
import { useFeeds } from "@/hooks/use-feeds";
import { useUIStore } from "@/store";
import { cn } from "@/lib/utils";

export function FeedList() {
  const { groups, feeds, isLoading, getFeedsByGroup, getUnreadCount, getGroupUnreadCount, getTotalUnreadCount } =
    useFeeds();
  const { selectedFeedId, selectedGroupId, selectAll } = useUIStore();

  const isAllSelected = selectedFeedId === null && selectedGroupId === null;
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
    <ScrollArea className="flex-1">
      <div className="p-2 space-y-1">
        {/* All feeds */}
        <button
          onClick={selectAll}
          className={cn(
            "flex w-full items-center gap-2.5 rounded-md px-2 py-1.5 text-left text-sm transition-colors",
            isAllSelected
              ? "bg-accent text-accent-foreground"
              : "hover:bg-accent/50"
          )}
        >
          <Inbox className="h-4 w-4 shrink-0 text-muted-foreground" />
          <span className="flex-1">All</span>
          {totalUnread > 0 && (
            <span className="text-xs text-muted-foreground">{totalUnread}</span>
          )}
        </button>

        {/* Feed groups */}
        <div className="mt-2 space-y-1">
          {groups.map((group) => {
            const groupFeeds = getFeedsByGroup(group.id);
            if (groupFeeds.length === 0) return null;

            return (
              <FeedGroup
                key={group.id}
                name={group.name}
                feeds={groupFeeds}
                unreadCount={getGroupUnreadCount(group.id)}
                getUnreadCount={getUnreadCount}
              />
            );
          })}

          {/* Ungrouped feeds (group_id = 0) */}
          {feeds
            .filter((f) => f.group_id === 0)
            .map((feed) => (
              <div key={feed.id} className="pl-5">
                <button
                  onClick={() => useUIStore.getState().setSelectedFeed(feed.id)}
                  className={cn(
                    "flex w-full items-center gap-2.5 rounded-md px-2 py-1.5 text-left text-sm transition-colors",
                    selectedFeedId === feed.id
                      ? "bg-accent text-accent-foreground"
                      : "hover:bg-accent/50"
                  )}
                >
                  <div
                    className="h-4 w-4 shrink-0 rounded"
                    style={{ backgroundColor: "#EB5757" }}
                  />
                  <span className="flex-1 truncate">{feed.name}</span>
                  {getUnreadCount(feed.id) > 0 && (
                    <span className="text-xs text-muted-foreground">
                      {getUnreadCount(feed.id)}
                    </span>
                  )}
                </button>
              </div>
            ))}
        </div>
      </div>
    </ScrollArea>
  );
}
