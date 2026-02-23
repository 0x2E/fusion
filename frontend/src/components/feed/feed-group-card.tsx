import { AlertCircle, ChevronDown, ChevronRight, Folder, Pause, Pencil, Plus, Trash2 } from "lucide-react";
import type { Dispatch, SetStateAction } from "react";
import { FeedFavicon } from "@/components/feed/feed-favicon";
import { Input } from "@/components/ui/input";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { getFaviconUrl } from "@/lib/api/favicon";
import { getFeedErrorPreview } from "@/lib/feed-error";
import type { TranslationKey } from "@/lib/i18n";
import { cn, formatDate } from "@/lib/utils";
import type { Feed, Group } from "@/lib/api";

interface FeedGroupCardProps {
  group: Group;
  groupFeeds: Feed[];
  isCollapsed: boolean;
  isEditing: boolean;
  editingGroupName: string;
  isMobile: boolean;
  mobileErrorTooltipFeedId: number | null;
  onToggleGroup: (groupId: number) => void;
  onStartEditingGroup: (group: Group) => void;
  onChangeEditingGroupName: (value: string) => void;
  onSaveGroupName: (group: Group) => void;
  onCancelEditingGroup: () => void;
  onOpenAddFeed: () => void;
  onOpenDeleteGroup: (group: Group) => void;
  onOpenEditFeed: (feed: Feed) => void;
  onChangeMobileErrorTooltipFeedId: Dispatch<SetStateAction<number | null>>;
  t: (key: TranslationKey, params?: Record<string, string | number>) => string;
}

function getDomain(url: string) {
  try {
    return new URL(url).hostname;
  } catch {
    return url;
  }
}

export function FeedGroupCard({
  group,
  groupFeeds,
  isCollapsed,
  isEditing,
  editingGroupName,
  isMobile,
  mobileErrorTooltipFeedId,
  onToggleGroup,
  onStartEditingGroup,
  onChangeEditingGroupName,
  onSaveGroupName,
  onCancelEditingGroup,
  onOpenAddFeed,
  onOpenDeleteGroup,
  onOpenEditFeed,
  onChangeMobileErrorTooltipFeedId,
  t,
}: FeedGroupCardProps) {
  return (
    <div className="overflow-hidden rounded-lg border">
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
              onChange={(e) => onChangeEditingGroupName(e.target.value)}
              onBlur={() => onSaveGroupName(group)}
              onKeyDown={(e) => {
                if (e.key === "Enter") onSaveGroupName(group);
                if (e.key === "Escape") onCancelEditingGroup();
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
            onClick={() => onToggleGroup(group.id)}
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
              onOpenAddFeed();
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
              onStartEditingGroup(group);
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
                onOpenDeleteGroup(group);
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
                index < groupFeeds.length - 1 && "border-b border-border/50",
              )}
            >
              <div className="flex min-w-0 flex-1 items-center gap-2.5">
                <FeedFavicon
                  src={getFaviconUrl(feed.link, feed.site_url)}
                  className="h-5 w-5"
                />
                <div className="min-w-0">
                  <p className="truncate text-[13px] font-medium">{feed.name}</p>
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
                      onChangeMobileErrorTooltipFeedId((current) =>
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
                          onChangeMobileErrorTooltipFeedId((current) =>
                            current === feed.id ? null : feed.id,
                          );
                        }}
                        className="flex items-center gap-1 rounded-sm text-xs text-destructive"
                      >
                        <AlertCircle className="h-3.5 w-3.5 shrink-0" />
                        <span className="hidden max-w-56 truncate font-medium sm:inline">
                          {getFeedErrorPreview(feed.fetch_state.last_error)}
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
                    ? formatDate(feed.fetch_state.last_checked_at)
                    : t("common.unknown")}
                </span>
                <button
                  type="button"
                  onClick={() => onOpenEditFeed(feed)}
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
}
