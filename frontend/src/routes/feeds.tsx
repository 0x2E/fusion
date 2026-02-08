import { useState, useMemo } from "react";
import { createFileRoute } from "@tanstack/react-router";
import {
  ChevronDown,
  ChevronRight,
  Download,
  Folder,
  ListFilter,
  Pause,
  Pencil,
  Plus,
  RefreshCw,
  Rss,
  Search,
  Trash2,
  Upload,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { AppLayout } from "@/components/layout/app-layout";
import { SidebarTrigger } from "@/components/layout/sidebar-trigger";
import { useFeeds } from "@/hooks/use-feeds";
import { useUIStore, useDataStore } from "@/store";
import { feedAPI, groupAPI } from "@/lib/api";
import type { Feed, Group } from "@/lib/api";
import { getFaviconUrl } from "@/lib/api/favicon";
import { generateOPML, downloadFile } from "@/lib/opml";
import { toast } from "sonner";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/feeds")({
  component: FeedsPage,
});

type StatusFilter = "all" | "error" | "paused";

const statusFilterLabels: Record<StatusFilter, string> = {
  all: "All Status",
  error: "Error",
  paused: "Paused",
};

function FeedsPage() {
  const { feeds, groups, getFeedsByGroup } = useFeeds();
  const {
    setEditFeedOpen,
    setImportOpmlOpen,
    setAddFeedOpen,
    setAddGroupOpen,
  } = useUIStore();
  const { updateGroup, removeGroup, moveFeedsToGroup } = useDataStore();

  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isExporting, setIsExporting] = useState(false);
  const [collapsedGroups, setCollapsedGroups] = useState<Set<number>>(
    new Set(),
  );

  // Group inline rename
  const [editingGroupId, setEditingGroupId] = useState<number | null>(null);
  const [editingGroupName, setEditingGroupName] = useState("");

  // Group delete confirmation
  const [deletingGroup, setDeletingGroup] = useState<Group | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  // Refresh all confirmation
  const [refreshConfirmOpen, setRefreshConfirmOpen] = useState(false);

  const isFiltering = searchQuery.trim() !== "" || statusFilter !== "all";

  const groupedFeeds = useMemo(() => {
    const query = searchQuery.toLowerCase().trim();

    const matchesFeed = (feed: Feed) => {
      if (
        query &&
        !feed.name.toLowerCase().includes(query) &&
        !feed.link.toLowerCase().includes(query)
      ) {
        return false;
      }
      if (statusFilter === "error" && !feed.failure) return false;
      if (statusFilter === "paused" && !feed.suspended) return false;
      return true;
    };

    return groups.map((group) => ({
      group,
      feeds: getFeedsByGroup(group.id).filter(matchesFeed),
    }));
  }, [groups, feeds, searchQuery, statusFilter, getFeedsByGroup]);

  const toggleGroup = (groupId: number) => {
    setCollapsedGroups((prev) => {
      const next = new Set(prev);
      if (next.has(groupId)) {
        next.delete(groupId);
      } else {
        next.add(groupId);
      }
      return next;
    });
  };

  const handleRefreshAll = async () => {
    setIsRefreshing(true);
    try {
      await feedAPI.refresh();
      toast.success("Refreshing all feeds...");
    } catch {
      toast.error("Failed to refresh feeds");
    } finally {
      setIsRefreshing(false);
    }
  };

  const handleExport = async () => {
    setIsExporting(true);
    try {
      const [groupsRes, feedsRes] = await Promise.all([
        groupAPI.list(),
        feedAPI.list(),
      ]);
      const opml = generateOPML(groupsRes.data, feedsRes.data);
      downloadFile(opml, "fusion-subscriptions.opml", "application/xml");
      toast.success("OPML exported successfully");
    } catch {
      toast.error("Failed to export OPML");
    } finally {
      setIsExporting(false);
    }
  };

  const startEditingGroup = (group: Group) => {
    setEditingGroupId(group.id);
    setEditingGroupName(group.name);
  };

  const saveGroupName = async (group: Group) => {
    const name = editingGroupName.trim();
    setEditingGroupId(null);
    if (!name || name === group.name) return;

    try {
      await groupAPI.update(group.id, { name });
      updateGroup(group.id, name);
      toast.success("Group renamed");
    } catch {
      toast.error("Failed to rename group");
    }
  };

  const confirmDeleteGroup = async () => {
    if (!deletingGroup) return;

    setIsDeleting(true);
    try {
      const groupFeeds = feeds.filter((f) => f.group_id === deletingGroup.id);

      // Move feeds to default group (id=1)
      await Promise.all(
        groupFeeds.map((feed) => feedAPI.update(feed.id, { group_id: 1 })),
      );
      await groupAPI.delete(deletingGroup.id);

      moveFeedsToGroup(deletingGroup.id, 1);
      removeGroup(deletingGroup.id);
      toast.success("Group deleted");
      setDeletingGroup(null);
    } catch {
      toast.error("Failed to delete group");
    } finally {
      setIsDeleting(false);
    }
  };

  const getDomain = (url: string) => {
    try {
      return new URL(url).hostname;
    } catch {
      return url;
    }
  };

  const visibleGroups = groupedFeeds.filter(
    ({ feeds: gf }) => gf.length > 0 || !isFiltering,
  );

  const totalVisible = groupedFeeds.reduce((sum, g) => sum + g.feeds.length, 0);

  return (
    <AppLayout>
      <div className="flex h-full flex-col">
        {/* Header */}
        <header className="flex items-center justify-between border-b px-4 py-4 sm:px-6">
          <div className="flex items-center gap-1">
            <SidebarTrigger />
            <h1 className="text-lg font-semibold">Manage Feeds</h1>
          </div>
          <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
            <Rss className="h-4 w-4" />
            <span className="font-medium">
              {feeds.length} {feeds.length === 1 ? "feed" : "feeds"}
            </span>
          </div>
        </header>

        {/* Toolbar */}
        <div className="flex flex-col gap-3 px-4 py-3 sm:flex-row sm:items-center sm:justify-between sm:px-6">
          <div className="flex items-center gap-2">
            <div className="relative flex-1 sm:flex-initial">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Search feeds..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="h-9 w-full pl-9 sm:w-[280px]"
              />
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="outline"
                  size="sm"
                  className="shrink-0 gap-1.5"
                >
                  <ListFilter className="h-3.5 w-3.5" />
                  <span className="hidden sm:inline">
                    {statusFilterLabels[statusFilter]}
                  </span>
                  <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start">
                {(Object.keys(statusFilterLabels) as StatusFilter[]).map(
                  (key) => (
                    <DropdownMenuItem
                      key={key}
                      onSelect={() => setStatusFilter(key)}
                    >
                      {statusFilterLabels[key]}
                    </DropdownMenuItem>
                  ),
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
          <div className="flex items-center gap-2">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button size="sm">
                  <Plus className="mr-1.5 h-3.5 w-3.5" />
                  Add
                  <ChevronDown className="ml-1 h-3.5 w-3.5" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onSelect={() => setAddFeedOpen(true)}>
                  <Rss className="mr-2 h-4 w-4" />
                  Add Feed
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={() => setAddGroupOpen(true)}>
                  <Folder className="mr-2 h-4 w-4" />
                  Add Group
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setRefreshConfirmOpen(true)}
              disabled={isRefreshing}
            >
              <RefreshCw
                className={cn(
                  "h-3.5 w-3.5 sm:mr-1.5",
                  isRefreshing && "animate-spin",
                )}
              />
              <span className="hidden sm:inline">Refresh All</span>
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setImportOpmlOpen(true)}
            >
              <Upload className="h-3.5 w-3.5 sm:mr-1.5" />
              <span className="hidden sm:inline">Import</span>
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleExport}
              disabled={isExporting}
            >
              <Download className="h-3.5 w-3.5 sm:mr-1.5" />
              <span className="hidden sm:inline">
                {isExporting ? "Exporting..." : "Export"}
              </span>
            </Button>
          </div>
        </div>

        {/* Collapsible Groups */}
        <ScrollArea className="flex-1">
          <div className="space-y-2 p-4 sm:p-6">
            {visibleGroups.map(({ group, feeds: groupFeeds }) => {
              const isCollapsed = collapsedGroups.has(group.id);
              const isEditing = editingGroupId === group.id;

              return (
                <div
                  key={group.id}
                  className="overflow-hidden rounded-lg border"
                >
                  {/* Group Header */}
                  <div
                    onClick={() => {
                      if (!isEditing) toggleGroup(group.id);
                    }}
                    className={cn(
                      "group/header flex cursor-pointer items-center justify-between bg-muted/50 px-3.5 py-2.5",
                      isCollapsed ? "rounded-lg" : "rounded-t-lg",
                    )}
                  >
                    <div className="flex items-center gap-2">
                      {isCollapsed ? (
                        <ChevronRight className="h-4 w-4 text-muted-foreground" />
                      ) : (
                        <ChevronDown className="h-4 w-4" />
                      )}
                      <Folder
                        className={cn(
                          "h-4 w-4",
                          isCollapsed && "text-muted-foreground",
                        )}
                      />
                      {isEditing ? (
                        <Input
                          value={editingGroupName}
                          onChange={(e) => setEditingGroupName(e.target.value)}
                          onBlur={() => saveGroupName(group)}
                          onKeyDown={(e) => {
                            if (e.key === "Enter") saveGroupName(group);
                            if (e.key === "Escape") setEditingGroupId(null);
                          }}
                          onClick={(e) => e.stopPropagation()}
                          className="h-7 w-40 px-2 text-sm"
                          autoFocus
                        />
                      ) : (
                        <span
                          className={cn(
                            "text-sm",
                            isCollapsed ? "font-medium" : "font-semibold",
                          )}
                        >
                          {group.name}
                        </span>
                      )}
                      <span
                        className={cn(
                          "rounded-full bg-muted px-2 py-0.5 text-[11px]",
                          isCollapsed
                            ? "font-medium text-muted-foreground"
                            : "font-semibold text-muted-foreground",
                        )}
                      >
                        {groupFeeds.length}
                      </span>
                    </div>
                    <div className="flex items-center gap-1 sm:opacity-0 sm:transition-opacity sm:group-hover/header:opacity-100">
                      <button
                        type="button"
                        onClick={(e) => {
                          e.stopPropagation();
                          setAddFeedOpen(true);
                        }}
                        className="rounded p-1 hover:bg-accent"
                      >
                        <Plus className="h-3.5 w-3.5 text-muted-foreground" />
                      </button>
                      <button
                        type="button"
                        onClick={(e) => {
                          e.stopPropagation();
                          startEditingGroup(group);
                        }}
                        className="rounded p-1 hover:bg-accent"
                      >
                        <Pencil className="h-3.5 w-3.5 text-muted-foreground" />
                      </button>
                      {group.id !== 1 && (
                        <button
                          type="button"
                          onClick={(e) => {
                            e.stopPropagation();
                            setDeletingGroup(group);
                          }}
                          className="rounded p-1 hover:bg-accent"
                        >
                          <Trash2 className="h-3.5 w-3.5 text-muted-foreground" />
                        </button>
                      )}
                    </div>
                  </div>

                  {/* Feed List */}
                  {!isCollapsed && groupFeeds.length > 0 && (
                    <div>
                      {groupFeeds.map((feed, index) => (
                        <div
                          key={feed.id}
                          className={cn(
                            "flex items-center justify-between py-2.5 pl-8 pr-3.5 transition-colors hover:bg-accent/30 sm:pl-11",
                            index < groupFeeds.length - 1 &&
                              "border-b border-border/50",
                          )}
                        >
                          <div className="flex min-w-0 flex-1 items-center gap-2.5">
                            <img
                              src={getFaviconUrl(feed.link, feed.site_url)}
                              alt=""
                              className="h-5 w-5 shrink-0 rounded"
                              loading="lazy"
                            />
                            <div className="min-w-0">
                              <p className="truncate text-[13px] font-medium">
                                {feed.name}
                              </p>
                              <p className="truncate text-[11px] text-muted-foreground">
                                {getDomain(feed.link)}
                              </p>
                            </div>
                          </div>
                          <div className="flex shrink-0 items-center gap-1.5 sm:gap-2.5">
                            {feed.failure && (
                              <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-destructive" />
                            )}
                            {feed.suspended && (
                              <Pause className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                            )}
                            <span className="hidden text-xs text-muted-foreground sm:inline">
                              {feed.item_count} articles
                            </span>
                            <button
                              type="button"
                              onClick={() => setEditFeedOpen(true, feed)}
                              className="rounded p-1 hover:bg-accent"
                            >
                              <Pencil className="h-3.5 w-3.5 text-muted-foreground" />
                            </button>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              );
            })}

            {visibleGroups.length === 0 && (
              <div className="py-12 text-center text-sm text-muted-foreground">
                {isFiltering ? "No feeds match your filters" : "No feeds yet"}
              </div>
            )}

            {isFiltering && visibleGroups.length > 0 && totalVisible === 0 && (
              <div className="py-12 text-center text-sm text-muted-foreground">
                No feeds match your filters
              </div>
            )}
          </div>
        </ScrollArea>
      </div>

      {/* Delete Group Confirmation Dialog */}
      <Dialog
        open={deletingGroup !== null}
        onOpenChange={(open) => !open && setDeletingGroup(null)}
      >
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Delete Group</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete{" "}
              <span className="font-semibold">{deletingGroup?.name}</span>?
              {(() => {
                const count = feeds.filter(
                  (f) => f.group_id === deletingGroup?.id,
                ).length;
                if (count === 0) return "";
                const target = groups.find((g) => g.id === 1);
                return (
                  <>
                    {" "}
                    All {count} feed(s) will be moved to{" "}
                    <span className="font-semibold">
                      {target?.name ?? "Default"}
                    </span>
                    .
                  </>
                );
              })()}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeletingGroup(null)}
              disabled={isDeleting}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={confirmDeleteGroup}
              disabled={isDeleting}
            >
              {isDeleting ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      {/* Refresh All Confirmation Dialog */}
      <Dialog open={refreshConfirmOpen} onOpenChange={setRefreshConfirmOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Refresh All Feeds</DialogTitle>
            <DialogDescription>
              This will refresh all {feeds.length} feeds. Continue?
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setRefreshConfirmOpen(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={() => {
                setRefreshConfirmOpen(false);
                handleRefreshAll();
              }}
            >
              Refresh
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </AppLayout>
  );
}
