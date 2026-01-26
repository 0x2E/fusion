import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useFeeds } from "@/hooks/use-feeds";
import { cn } from "@/lib/utils";
import { useUIStore } from "@/store";
import { FolderPlus, Layers, Plus } from "lucide-react";
import { FeedGroup } from "./feed-group";

export function FeedList() {
  const {
    groups,
    feeds,
    isLoading,
    getFeedsByGroup,
    getUnreadCount,
    getGroupUnreadCount,
    getTotalUnreadCount,
  } = useFeeds();
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
      <div className="p-2 space-y-0.5">
        {/* Feeds header */}
        <div className="flex items-center justify-between px-2 py-1">
          <span className="text-xs font-medium text-muted-foreground">
            Feeds
          </span>
          <div className="flex items-center gap-0.5">
            <Button variant="ghost" size="icon" className="h-5 w-5">
              <FolderPlus className="h-3.5 w-3.5 text-muted-foreground" />
            </Button>
            <Button variant="ghost" size="icon" className="h-5 w-5">
              <Plus className="h-3.5 w-3.5 text-muted-foreground" />
            </Button>
          </div>
        </div>

        {/* All feeds */}
        <button
          onClick={selectAll}
          className={cn(
            "flex w-full items-center gap-2.5 rounded-md px-2 py-1.5 text-left text-sm transition-colors",
            isAllSelected
              ? "bg-accent text-accent-foreground"
              : "hover:bg-accent/50",
          )}
        >
          <Layers className="h-4 w-4 shrink-0 text-muted-foreground" />
          <span className="flex-1">All</span>
          {totalUnread > 0 && (
            <span className="text-xs text-muted-foreground">{totalUnread}</span>
          )}
        </button>

        {/* Feed groups */}
        <div className="mt-2 space-y-1">
          {groups.map((group) => {
            const groupFeeds = getFeedsByGroup(group.id);

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
                      : "hover:bg-accent/50",
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
