import { useMemo } from "react";
import { useLocation } from "@tanstack/react-router";
import { Inbox, Layers, Star } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { isArticleFilter, type ArticleFilter } from "@/lib/article-filter";
import { useGroups } from "@/queries/groups";
import { useFeedLookup, useUnreadCounts } from "@/queries/feeds";
import { useBookmarkLookup } from "@/queries/bookmarks";
import { useUrlState } from "@/hooks/use-url-state";
import { useSidebarNavigation } from "@/hooks/use-keyboard";
import { useI18n } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import { sidebarNodeKey, useSidebarNavStore, type SidebarNode } from "@/store";
import { FeedGroup } from "./feed-group";
import { FeedItem } from "./feed-item";

const sidebarFilterValues: ArticleFilter[] = ["unread", "starred", "all"];

export function FeedList() {
  const { t } = useI18n();
  const { data: groups = [], isLoading } = useGroups();
  const { feeds, getFeedsByGroup } = useFeedLookup();
  const { getTotalUnreadCount } = useUnreadCounts();
  const { total: starredTotal } = useBookmarkLookup();
  const {
    selectedFeedId,
    selectedGroupId,
    articleFilter,
    selectTopLevelFilter,
  } = useUrlState();
  const { pathname } = useLocation();
  const collapsedGroupIds = useSidebarNavStore((s) => s.collapsedGroupIds);
  const focusedKey = useSidebarNavStore((s) => s.focusedKey);
  const setFocusedKey = useSidebarNavStore((s) => s.setFocusedKey);
  const firstPathSegment = pathname.split("/").filter(Boolean)[0];
  const isOnHomePage =
    typeof firstPathSegment === "string" && isArticleFilter(firstPathSegment);
  const isTopLevelSelected =
    isOnHomePage && selectedFeedId === null && selectedGroupId === null;
  const totalUnread = getTotalUnreadCount();
  const starredCount = starredTotal;

  const topFilters: Array<{
    value: "all" | "unread" | "starred";
    label: string;
    count: number;
    icon: typeof Inbox;
  }> = [
    {
      value: "unread",
      label: t("article.filter.unread"),
      count: totalUnread,
      icon: Inbox,
    },
    {
      value: "starred",
      label: t("article.filter.starred"),
      count: starredCount,
      icon: Star,
    },
    {
      value: "all",
      label: t("article.filter.all"),
      count: totalUnread,
      icon: Layers,
    },
  ];

  // Flat, ordered view of the tree used by Shift+N / Shift+P keyboard navigation.
  const nodes = useMemo<SidebarNode[]>(() => {
    const list: SidebarNode[] = sidebarFilterValues.map((filter) => ({
      key: sidebarNodeKey.filter(filter),
      kind: "filter",
      filter,
    }));

    for (const group of groups) {
      list.push({
        key: sidebarNodeKey.group(group.id),
        kind: "group",
        groupId: group.id,
      });

      if (!collapsedGroupIds.includes(group.id)) {
        for (const feed of getFeedsByGroup(group.id)) {
          list.push({
            key: sidebarNodeKey.feed(feed.id),
            kind: "feed",
            feedId: feed.id,
            groupId: group.id,
          });
        }
      }
    }

    for (const feed of feeds.filter((f) => f.group_id === 0)) {
      list.push({
        key: sidebarNodeKey.feed(feed.id),
        kind: "feed",
        feedId: feed.id,
        groupId: 0,
      });
    }

    return list;
  }, [collapsedGroupIds, feeds, getFeedsByGroup, groups]);

  useSidebarNavigation(nodes);

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
    <ScrollArea className="min-h-0 flex-1 w-full min-w-0 overflow-hidden [&_[data-slot=scroll-area-viewport]>div]:!block">
      <div className="w-full min-w-0 p-2 space-y-0.5">
        {/* Top-level filters */}
        <div className="space-y-0.5">
          {topFilters.map(({ value, label, count, icon: Icon }) => {
            const nodeKey = sidebarNodeKey.filter(value);

            return (
              <button
                key={value}
                data-sidebar-key={nodeKey}
                onClick={() => {
                  setFocusedKey(nodeKey);
                  selectTopLevelFilter(value);
                }}
                className={cn(
                  "flex w-full min-w-0 items-center gap-1.5 rounded-md px-2 py-1 text-left text-sm transition-colors",
                  isTopLevelSelected && articleFilter === value
                    ? "bg-accent text-accent-foreground"
                    : "hover:bg-accent/50",
                  focusedKey === nodeKey && "ring-2 ring-inset ring-ring",
                )}
              >
                <Icon className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                <span className="min-w-0 flex-1">{label}</span>
                <span className="shrink-0 text-[11px] text-muted-foreground">
                  {count}
                </span>
              </button>
            );
          })}
        </div>

        {/* Feeds header */}
        <div className="mt-2 flex items-center justify-between px-2 py-1">
          <span className="text-[11px] font-medium text-muted-foreground">
            {t("search.group.feeds")}
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
