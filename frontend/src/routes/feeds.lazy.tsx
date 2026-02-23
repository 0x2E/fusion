import { useMemo, useState } from "react";
import { createLazyFileRoute } from "@tanstack/react-router";
import {
  ChevronDown,
  Download,
  Folder,
  ListFilter,
  Plus,
  RefreshCw,
  Rss,
  Search,
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
import { feedAPI, groupAPI } from "@/lib/api";
import type { Feed, Group } from "@/lib/api";
import { generateOPML, downloadFile } from "@/lib/opml";
import { useI18n } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import {
  useFeedLookup,
  useMoveFeedsToGroup,
  useRefreshFeeds,
} from "@/queries/feeds";
import { useDeleteGroup, useGroups, useUpdateGroup } from "@/queries/groups";
import { useUIStore } from "@/store";
import { FeedGroupCard } from "@/components/feed/feed-group-card";
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

  const visibleGroups = groupedFeeds.filter(
    ({ feeds: gf }) => gf.length > 0 || !isFiltering,
  );

  const totalVisible = groupedFeeds.reduce((sum, g) => sum + g.feeds.length, 0);
  const hasNoFeeds = !isFeedsLoading && feeds.length === 0;
  const deletingGroupMoveHint = useMemo(() => {
    if (!deletingGroup) {
      return "";
    }

    const count = feeds.filter((feed) => feed.group_id === deletingGroup.id).length;
    if (count === 0) {
      return "";
    }

    const target = groups.find((group) => group.id === 1);
    return t("feeds.deleteGroup.moveHint", {
      count,
      target: target?.name ?? t("feeds.deleteGroup.targetDefault"),
    });
  }, [deletingGroup, feeds, groups, t]);

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
                    <FeedGroupCard
                      key={group.id}
                      group={group}
                      groupFeeds={groupFeeds}
                      isCollapsed={isCollapsed}
                      isEditing={isEditing}
                      editingGroupName={editingGroupName}
                      isMobile={isMobile}
                      mobileErrorTooltipFeedId={mobileErrorTooltipFeedId}
                      onToggleGroup={toggleGroup}
                      onStartEditingGroup={startEditingGroup}
                      onChangeEditingGroupName={setEditingGroupName}
                      onSaveGroupName={(targetGroup) => {
                        void saveGroupName(targetGroup);
                      }}
                      onCancelEditingGroup={() => setEditingGroupId(null)}
                      onOpenAddFeed={() => setAddFeedOpen(true)}
                      onOpenDeleteGroup={setDeletingGroup}
                      onOpenEditFeed={(feed) => setEditFeedOpen(true, feed)}
                      onChangeMobileErrorTooltipFeedId={setMobileErrorTooltipFeedId}
                      t={t}
                    />
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
              {deletingGroupMoveHint ? ` ${deletingGroupMoveHint}` : ""}
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
