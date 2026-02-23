import { useMemo, useState } from "react";
import { createLazyFileRoute } from "@tanstack/react-router";
import {
  AlertCircle,
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
import { toast } from "sonner";
import { AppLayout } from "@/components/layout/app-layout";
import { ContentHeader } from "@/components/layout/content-header";
import { SidebarTrigger } from "@/components/layout/sidebar-trigger";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { feedAPI, groupAPI } from "@/lib/api";
import type { Feed, Group } from "@/lib/api";
import { getFeedErrorPreview } from "@/lib/feed-error";
import { getFaviconUrl } from "@/lib/api/favicon";
import { generateOPML, downloadFile } from "@/lib/opml";
import { useI18n } from "@/lib/i18n";
import { cn, formatDate } from "@/lib/utils";
import {
  useFeedLookup,
  useMoveFeedsToGroup,
  useRefreshFeeds,
} from "@/queries/feeds";
import { useDeleteGroup, useGroups, useUpdateGroup } from "@/queries/groups";
import { useUIStore } from "@/store";
import { FeedFavicon } from "@/components/feed/feed-favicon";
import { useIsMobile } from "@/hooks/use-mobile";

export const Route = createLazyFileRoute("/feeds")({
  component: FeedsPage,
});

type StatusFilter = "all" | "error" | "paused";

function FeedsPage() {
  const { t } = useI18n();
  const { data: groups = [] } = useGroups();
  const { feeds, getFeedsByGroup, isLoading: isFeedsLoading } = useFeedLookup();
  const updateGroupMutation = useUpdateGroup();
  const deleteGroupMutation = useDeleteGroup();
  const moveFeedsMutation = useMoveFeedsToGroup();
  const refreshFeedsMutation = useRefreshFeeds();

  const {
    setEditFeedOpen,
    setImportOpmlOpen,
    setAddFeedOpen,
    setAddGroupOpen,
  } = useUIStore();

  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [isExporting, setIsExporting] = useState(false);
  const [collapsedGroups, setCollapsedGroups] = useState<Set<number>>(
    new Set(),
  );

  const [editingGroupId, setEditingGroupId] = useState<number | null>(null);
  const [editingGroupName, setEditingGroupName] = useState("");

  const [deletingGroup, setDeletingGroup] = useState<Group | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const [refreshConfirmOpen, setRefreshConfirmOpen] = useState(false);
  const [mobileErrorTooltipFeedId, setMobileErrorTooltipFeedId] = useState<
    number | null
  >(null);
  const isMobile = useIsMobile();

  const statusFilterLabels: Record<StatusFilter, string> = {
    all: t("feeds.status.all"),
    error: t("feeds.status.error"),
    paused: t("feeds.status.paused"),
  };

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
      if (statusFilter === "error" && !feed.fetch_state.last_error)
        return false;
      if (statusFilter === "paused" && !feed.suspended) return false;
      return true;
    };

    return groups.map((group) => ({
      group,
      feeds: getFeedsByGroup(group.id).filter(matchesFeed),
    }));
  }, [groups, searchQuery, statusFilter, getFeedsByGroup]);

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
    try {
      await refreshFeedsMutation.mutateAsync();
      toast.success(t("feeds.toast.refreshing"));
    } catch {
      toast.error(t("feeds.toast.refreshFailed"));
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
      toast.success(t("feeds.toast.exported"));
    } catch {
      toast.error(t("feeds.toast.exportFailed"));
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
      await updateGroupMutation.mutateAsync({ id: group.id, name });
      toast.success(t("feeds.toast.renamed"));
    } catch {
      toast.error(t("feeds.toast.renameFailed"));
    }
  };

  const confirmDeleteGroup = async () => {
    if (!deletingGroup) return;

    setIsDeleting(true);
    try {
      await moveFeedsMutation.mutateAsync({
        fromGroupId: deletingGroup.id,
        toGroupId: 1,
      });
      await deleteGroupMutation.mutateAsync(deletingGroup.id);

      toast.success(t("feeds.toast.deleted"));
      setDeletingGroup(null);
    } catch {
      toast.error(t("feeds.toast.deleteFailed"));
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
  const hasNoFeeds = !isFeedsLoading && feeds.length === 0;

  return (
    <AppLayout>
      <div className="flex h-full min-h-0 flex-col overflow-hidden">
        <ContentHeader>
          <div className="flex items-center gap-1">
            <SidebarTrigger />
            <h1 className="text-lg font-semibold">{t("feeds.header")}</h1>
          </div>
          <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
            <Rss className="h-4 w-4" />
            <span className="font-medium">
              {t("feeds.count", { count: feeds.length })}
            </span>
          </div>
        </ContentHeader>

        <div className="flex flex-col gap-3 px-4 py-3 sm:flex-row sm:items-center sm:justify-between sm:px-6">
          <div className="flex items-center gap-2">
            <div className="relative flex-1 sm:flex-initial">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder={t("feeds.searchPlaceholder")}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="h-9 w-full pl-9 sm:w-[280px]"
                name="feed-search"
                autoComplete="off"
                aria-label={t("feeds.searchPlaceholder")}
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
                  {t("common.add")}
                  <ChevronDown className="ml-1 h-3.5 w-3.5" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onSelect={() => setAddFeedOpen(true)}>
                  <Rss className="mr-2 h-4 w-4" />
                  {t("feed.add.title")}
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={() => setAddGroupOpen(true)}>
                  <Folder className="mr-2 h-4 w-4" />
                  {t("group.add.title")}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setRefreshConfirmOpen(true)}
              disabled={refreshFeedsMutation.isPending}
            >
              <RefreshCw
                className={cn(
                  "h-3.5 w-3.5 sm:mr-1.5",
                  refreshFeedsMutation.isPending && "animate-spin",
                )}
              />
              <span className="hidden sm:inline">{t("feeds.refreshAll")}</span>
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setImportOpmlOpen(true)}
            >
              <Upload className="h-3.5 w-3.5 sm:mr-1.5" />
              <span className="hidden sm:inline">{t("common.import")}</span>
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleExport}
              disabled={isExporting}
            >
              <Download className="h-3.5 w-3.5 sm:mr-1.5" />
              <span className="hidden sm:inline">
                {isExporting ? t("common.exporting") : t("feeds.exportButton")}
              </span>
            </Button>
          </div>
        </div>

        <ScrollArea className="min-h-0 flex-1">
          <div className="space-y-2 p-4 sm:p-6">
            {hasNoFeeds ? (
              <div className="py-12 text-center text-sm text-muted-foreground">
                {t("feeds.empty")}
              </div>
            ) : (
              <>
                {visibleGroups.map(({ group, feeds: groupFeeds }) => {
                  const isCollapsed = collapsedGroups.has(group.id);
                  const isEditing = editingGroupId === group.id;

                  return (
                    <div
                      key={group.id}
                      className="overflow-hidden rounded-lg border"
                    >
                      <div
                        className={cn(
                          "group/header flex items-center justify-between bg-muted/50 px-3.5 py-2.5",
                          isCollapsed ? "rounded-lg" : "rounded-t-lg",
                        )}
                      >
                        {isEditing ? (
                          <div className="flex min-w-0 flex-1 items-center gap-2 text-left">
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
                            <Input
                              value={editingGroupName}
                              onChange={(e) => setEditingGroupName(e.target.value)}
                              onBlur={() => saveGroupName(group)}
                              onKeyDown={(e) => {
                                if (e.key === "Enter") saveGroupName(group);
                                if (e.key === "Escape") setEditingGroupId(null);
                              }}
                              className="h-7 w-40 px-2 text-sm"
                              aria-label={t("group.add.placeholder")}
                              autoFocus={!isMobile}
                            />
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
                        ) : (
                          <button
                            type="button"
                            onClick={() => toggleGroup(group.id)}
                            className="flex min-w-0 flex-1 items-center gap-2 text-left"
                            aria-expanded={!isCollapsed}
                            aria-label={`Toggle group ${group.name}`}
                          >
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
                            <span
                              className={cn(
                                "text-sm",
                                isCollapsed ? "font-medium" : "font-semibold",
                              )}
                            >
                              {group.name}
                            </span>
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
                          </button>
                        )}
                        <div className="flex items-center gap-1 sm:opacity-0 sm:transition-opacity sm:group-hover/header:opacity-100">
                          <button
                            type="button"
                            onClick={(e) => {
                              e.stopPropagation();
                              setAddFeedOpen(true);
                            }}
                            className="rounded p-1 hover:bg-accent"
                            aria-label={t("feed.add.button")}
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
                            aria-label={t("feeds.toast.renamed")}
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
                              aria-label={t("feeds.deleteGroup.title")}
                            >
                              <Trash2 className="h-3.5 w-3.5 text-muted-foreground" />
                            </button>
                          )}
                        </div>
                      </div>

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
                                <FeedFavicon
                                  src={getFaviconUrl(feed.link, feed.site_url)}
                                  className="h-5 w-5"
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
                                {feed.fetch_state.last_error && (
                                  <Tooltip
                                    open={
                                      isMobile
                                        ? mobileErrorTooltipFeedId === feed.id
                                        : undefined
                                    }
                                    onOpenChange={(open) => {
                                      if (!isMobile || open) return;
                                      setMobileErrorTooltipFeedId((current) =>
                                        current === feed.id ? null : current,
                                      );
                                    }}
                                  >
                                    <TooltipTrigger asChild>
                                      <button
                                        type="button"
                                        aria-label={t("feeds.status.error")}
                                        onClick={() => {
                                          if (!isMobile) return;
                                          setMobileErrorTooltipFeedId((current) =>
                                            current === feed.id ? null : feed.id,
                                          );
                                        }}
                                        className="flex items-center gap-1 rounded-sm text-xs text-destructive"
                                      >
                                        <AlertCircle className="h-3.5 w-3.5 shrink-0" />
                                        <span className="hidden max-w-56 truncate font-medium sm:inline">
                                          {getFeedErrorPreview(
                                            feed.fetch_state.last_error,
                                          )}
                                        </span>
                                      </button>
                                    </TooltipTrigger>
                                    <TooltipContent
                                      side="top"
                                      className="max-w-sm whitespace-normal break-words"
                                    >
                                      {feed.fetch_state.last_error.trim()}
                                    </TooltipContent>
                                  </Tooltip>
                                )}
                                {feed.suspended && (
                                  <Pause className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                                )}
                                <span className="hidden text-xs text-muted-foreground sm:inline">
                                  {t("feeds.itemCount", {
                                    count: feed.item_count,
                                  })}{" "}
                                  ·{" "}
                                  {feed.fetch_state.last_checked_at > 0
                                    ? formatDate(
                                        feed.fetch_state.last_checked_at,
                                      )
                                    : t("common.unknown")}
                                </span>
                                <button
                                  type="button"
                                  onClick={() => setEditFeedOpen(true, feed)}
                                  className="rounded p-1 hover:bg-accent"
                                  aria-label={t("feed.edit.title")}
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
                    {t("feeds.noMatch")}
                  </div>
                )}

                {isFiltering &&
                  visibleGroups.length > 0 &&
                  totalVisible === 0 && (
                    <div className="py-12 text-center text-sm text-muted-foreground">
                      {t("feeds.noMatch")}
                    </div>
                  )}
              </>
            )}
          </div>
        </ScrollArea>
      </div>

      <Dialog
        open={deletingGroup !== null}
        onOpenChange={(open) => !open && setDeletingGroup(null)}
      >
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>{t("feeds.deleteGroup.title")}</DialogTitle>
            <DialogDescription>
              {t("feeds.deleteGroup.description", {
                name: deletingGroup?.name ?? "",
              })}
              {(() => {
                const count = feeds.filter(
                  (f) => f.group_id === deletingGroup?.id,
                ).length;
                if (count === 0) return "";
                const target = groups.find((g) => g.id === 1);
                return (
                  <>
                    {" "}
                    {t("feeds.deleteGroup.moveHint", {
                      count,
                      target:
                        target?.name ?? t("feeds.deleteGroup.targetDefault"),
                    })}
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
              {t("common.cancel")}
            </Button>
            <Button
              variant="destructive"
              onClick={confirmDeleteGroup}
              disabled={isDeleting}
            >
              {isDeleting ? t("common.deleting") : t("common.delete")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={refreshConfirmOpen} onOpenChange={setRefreshConfirmOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>{t("feeds.refreshDialog.title")}</DialogTitle>
            <DialogDescription>
              {t("feeds.refreshDialog.description", { count: feeds.length })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setRefreshConfirmOpen(false)}
            >
              {t("common.cancel")}
            </Button>
            <Button
              onClick={() => {
                setRefreshConfirmOpen(false);
                handleRefreshAll();
              }}
            >
              {t("common.refresh")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </AppLayout>
  );
}
