import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useFeeds } from "@/hooks/use-feeds";
import { cn } from "@/lib/utils";
import { useUIStore } from "@/store";
import { FolderPlus, Inbox, Plus, Settings, Upload } from "lucide-react";
import { FeedGroup } from "./feed-group";
import { FeedItem } from "./feed-item";

export function FeedList() {
  const { groups, feeds, isLoading, getFeedsByGroup, getTotalUnreadCount } =
    useFeeds();
  const {
    selectedFeedId,
    selectedGroupId,
    selectAll,
    setGroupManagementOpen,
    setAddFeedOpen,
    setFeedManagementOpen,
    setImportOpmlOpen,
  } = useUIStore();

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
    <ScrollArea className="flex-1 w-full min-w-0 overflow-hidden [&_[data-slot=scroll-area-viewport]>div]:!block">
      <div className="w-full min-w-0 p-2 space-y-0.5">
        {/* Feeds header */}
        <div className="flex items-center justify-between px-2 py-1">
          <span className="text-xs font-medium text-muted-foreground">
            Feeds
          </span>
          <div className="flex items-center gap-0.5">
            <Button
              variant="ghost"
              size="icon"
              className="h-5 w-5"
              onClick={() => setGroupManagementOpen(true)}
            >
              <FolderPlus className="h-3.5 w-3.5 text-muted-foreground" />
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon" className="h-5 w-5">
                  <Plus className="h-3.5 w-3.5 text-muted-foreground" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" className="w-[200px]">
                <DropdownMenuItem onSelect={() => setAddFeedOpen(true)}>
                  <Plus className="h-4 w-4" />
                  Add Feed
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={() => setImportOpmlOpen(true)}>
                  <Upload className="h-4 w-4" />
                  Import OPML
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem onSelect={() => setFeedManagementOpen(true)}>
                  <Settings className="h-4 w-4" />
                  Manage Feeds...
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {/* All feeds */}
        <button
          onClick={selectAll}
          className={cn(
            "flex w-full min-w-0 items-center gap-1.5 rounded-md px-2 py-1.5 text-left text-sm transition-colors",
            isAllSelected
              ? "bg-accent text-accent-foreground"
              : "hover:bg-accent/50",
          )}
        >
          <Inbox className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
          <span className="min-w-0 flex-1 font-medium">All</span>
          <span className="shrink-0 text-xs text-muted-foreground">
            {totalUnread}
          </span>
        </button>

        {/* Feed groups */}
        <div className="mt-2 w-full min-w-0 space-y-1">
          {groups.map((group) => {
            const groupFeeds = getFeedsByGroup(group.id);

            return (
              <FeedGroup key={group.id} name={group.name} feeds={groupFeeds} />
            );
          })}

          {/* Ungrouped feeds (group_id = 0) */}
          {feeds
            .filter((f) => f.group_id === 0)
            .map((feed) => (
              <FeedItem
                key={feed.id}
                id={feed.id}
                name={feed.name}
                feedLink={feed.link}
                siteUrl={feed.site_url}
                unreadCount={feed.unread_count}
              />
            ))}
        </div>
      </div>
    </ScrollArea>
  );
}
