import { FeedDialog } from "@/components/feed-dialog";
import { GroupDialog } from "@/components/group-dialog";
import { Button } from "@/components/ui/button";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import type { Feed, Group } from "@/lib/api/types";
import { cn } from "@/lib/utils";
import { useDataStore } from "@/store/data";
import { useNavigate, useSearch } from "@tanstack/react-router";
import {
  ChevronDown,
  ChevronRight,
  MoreHorizontal,
  Plus,
  Rows3,
  Rss,
  Search,
  Settings,
} from "lucide-react";
import { useState } from "react";

interface SidebarProps {
  onSearchOpen: () => void;
  onSettingsOpen: () => void;
}

export function Sidebar({ onSearchOpen, onSettingsOpen }: SidebarProps) {
  const navigate = useNavigate();
  const search = useSearch({ from: "/" });
  const {
    groups,
    feeds,
    groupsLoading,
    feedsLoading,
    deleteGroup,
    deleteFeed,
  } = useDataStore();
  const [openGroups, setOpenGroups] = useState<Record<number, boolean>>({});
  const [groupDialogOpen, setGroupDialogOpen] = useState(false);
  const [feedDialogOpen, setFeedDialogOpen] = useState(false);
  const [editingGroup, setEditingGroup] = useState<Group | undefined>();
  const [editingFeed, setEditingFeed] = useState<Feed | undefined>();
  const [feedDialogGroupId, setFeedDialogGroupId] = useState<
    number | undefined
  >();

  const toggleGroup = (groupId: number) => {
    setOpenGroups((prev) => ({ ...prev, [groupId]: !prev[groupId] }));
  };

  const handleSelectAll = () => {
    navigate({ to: "/", search: { filter: search.filter || "all" } });
  };

  const handleSelectFeed = (feedId: number) => {
    navigate({
      to: "/",
      search: { feed: feedId, filter: search.filter || "all" },
    });
  };

  const handleDeleteGroup = async (groupId: number) => {
    if (confirm("Are you sure you want to delete this group?")) {
      await deleteGroup(groupId);
      if (search.group === groupId) {
        navigate({ to: "/", search: { filter: search.filter || "all" } });
      }
    }
  };

  const handleDeleteFeed = async (feedId: number) => {
    if (confirm("Are you sure you want to unsubscribe from this feed?")) {
      await deleteFeed(feedId);
      if (search.feed === feedId) {
        navigate({ to: "/", search: { filter: search.filter || "all" } });
      }
    }
  };

  const handleCreateGroup = () => {
    setEditingGroup(undefined);
    setGroupDialogOpen(true);
  };

  const handleEditGroup = (group: Group) => {
    setEditingGroup(group);
    setGroupDialogOpen(true);
  };

  const handleCreateFeed = (groupId?: number) => {
    setEditingFeed(undefined);
    setFeedDialogGroupId(groupId);
    setFeedDialogOpen(true);
  };

  const handleEditFeed = (feed: Feed) => {
    setEditingFeed(feed);
    setFeedDialogGroupId(undefined);
    setFeedDialogOpen(true);
  };

  const getUnreadCount = (feedId: number) => {
    return feeds.find((f) => f.id === feedId)?.failures || 0;
  };

  const groupedFeeds = groups.map((group) => ({
    ...group,
    feeds: feeds.filter((f) => f.group_id === group.id),
  }));

  return (
    <aside className="w-64 h-full border-r border-border bg-sidebar flex flex-col">
      {/* Logo and Search */}
      <div className="p-4 space-y-3 border-b border-sidebar-border">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
            <Rss className="w-4 h-4 text-primary-foreground" />
          </div>
          <span className="font-semibold text-sidebar-foreground">
            RSS Reader
          </span>
        </div>

        <Button
          variant="ghost"
          className="w-full justify-start text-muted-foreground hover:text-foreground"
          onClick={onSearchOpen}
        >
          <Search className="w-4 h-4 mr-2" />
          Search
        </Button>
      </div>

      {/* Feed List */}
      <ScrollArea className="flex-1 p-2">
        <div className="space-y-0.5">
          <div className="px-2 py-1 flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              Feeds
            </span>
            <Button
              size="sm"
              variant="ghost"
              className="h-5 w-5 p-0"
              title="Create Group"
              onClick={handleCreateGroup}
            >
              <Plus className="w-3 h-3" />
            </Button>
          </div>

          {groupsLoading || feedsLoading ? (
            <div className="space-y-2 p-2">
              <Skeleton className="h-8 w-full" />
              <Skeleton className="h-8 w-full" />
              <Skeleton className="h-8 w-full" />
            </div>
          ) : (
            <>
              {/* All Feeds */}
              <Button
                variant="ghost"
                className={cn(
                  "w-full justify-start h-8 px-2 text-sm",
                  !search.feed &&
                    !search.group &&
                    "bg-sidebar-accent text-sidebar-accent-foreground"
                )}
                onClick={handleSelectAll}
              >
                <Rows3 className="w-3.5 h-3.5 mr-2" />
                <span>All</span>
              </Button>

              {/* Groups */}
              {groupedFeeds.map((group) => (
                <Collapsible
                  key={group.id}
                  open={openGroups[group.id]}
                  onOpenChange={() => toggleGroup(group.id)}
                >
                  <div className="group/group relative">
                    <CollapsibleTrigger asChild>
                      <Button
                        variant="ghost"
                        className="w-full justify-start h-8 px-2 pr-16 text-sm"
                      >
                        {openGroups[group.id] ? (
                          <ChevronDown className="w-3.5 h-3.5 mr-1.5" />
                        ) : (
                          <ChevronRight className="w-3.5 h-3.5 mr-1.5" />
                        )}
                        <span>{group.name}</span>
                      </Button>
                    </CollapsibleTrigger>

                    <div className="absolute right-0.5 top-0.5 opacity-0 group-hover/group:opacity-100 transition-opacity flex gap-0.5">
                      <Button
                        size="sm"
                        variant="ghost"
                        className="h-6 w-6 p-0"
                        title="Add Feed"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleCreateFeed(group.id);
                        }}
                      >
                        <Plus className="w-3 h-3" />
                      </Button>

                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            size="sm"
                            variant="ghost"
                            className="h-6 w-6 p-0"
                            onClick={(e) => e.stopPropagation()}
                          >
                            <MoreHorizontal className="w-3 h-3" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            onClick={() => handleCreateFeed(group.id)}
                          >
                            Add Feed
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => handleEditGroup(group)}
                          >
                            Rename Group
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            className="text-destructive"
                            onClick={() => handleDeleteGroup(group.id)}
                          >
                            Delete Group
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </div>

                  <CollapsibleContent className="ml-2 space-y-0.5 mt-0.5">
                    {group.feeds.map((feed) => (
                      <div
                        key={feed.id}
                        className="group/feed relative"
                        onContextMenu={(e) => {
                          e.preventDefault();
                        }}
                      >
                        <Button
                          variant="ghost"
                          className={cn(
                            "w-full justify-start h-8 pl-6 pr-2 relative",
                            search.feed === feed.id &&
                              "bg-sidebar-accent text-sidebar-accent-foreground"
                          )}
                          onClick={() => handleSelectFeed(feed.id)}
                        >
                          <span className="flex-1 text-left truncate text-sm">
                            {feed.name}
                          </span>
                          {getUnreadCount(feed.id) > 0 && (
                            <span className="absolute right-2 text-xs text-muted-foreground">
                              {getUnreadCount(feed.id)}
                            </span>
                          )}
                        </Button>

                        <div
                          className={cn(
                            "absolute right-0 top-0 h-full flex items-center pr-1 opacity-0 group-hover/feed:opacity-100 transition-opacity",
                            search.feed === feed.id
                              ? "bg-sidebar-accent"
                              : "bg-sidebar hover:bg-sidebar"
                          )}
                        >
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button
                                size="sm"
                                variant="ghost"
                                className="h-6 w-6 p-0 hover:bg-sidebar-accent"
                                onClick={(e) => e.stopPropagation()}
                              >
                                <MoreHorizontal className="w-3 h-3" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem>Mark as Read</DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => handleEditFeed(feed)}
                              >
                                Edit Feed
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                className="text-destructive"
                                onClick={() => handleDeleteFeed(feed.id)}
                              >
                                Unsubscribe
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </div>
                      </div>
                    ))}
                  </CollapsibleContent>
                </Collapsible>
              ))}
            </>
          )}
        </div>
      </ScrollArea>

      {/* Settings and Version */}
      <div className="p-2 border-t border-sidebar-border space-y-0.5">
        <Button
          variant="ghost"
          className="w-full justify-start h-8 px-2"
          onClick={onSettingsOpen}
        >
          <Settings className="w-4 h-4 mr-2" />
          Settings
        </Button>
        <div className="px-2 py-1 text-xs text-muted-foreground">v1.0.0</div>
      </div>

      <GroupDialog
        open={groupDialogOpen}
        onOpenChange={setGroupDialogOpen}
        group={editingGroup}
      />

      <FeedDialog
        open={feedDialogOpen}
        onOpenChange={setFeedDialogOpen}
        feed={editingFeed}
        defaultGroupId={feedDialogGroupId}
      />
    </aside>
  );
}
