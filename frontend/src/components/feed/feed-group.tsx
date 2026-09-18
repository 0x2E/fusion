import { ChevronRight } from "lucide-react";
import { Collapsible, CollapsibleContent } from "@/components/ui/collapsible";
import { cn } from "@/lib/utils";
import { useUrlState } from "@/hooks/use-url-state";
import { sidebarNodeKey, useSidebarNavStore } from "@/store";
import { FeedItem } from "./feed-item";
import type { Feed } from "@/lib/api";

interface FeedGroupProps {
  groupId: number;
  name: string;
  feeds: Feed[];
}

export function FeedGroup({ groupId, name, feeds }: FeedGroupProps) {
  const { selectedGroupId, setSelectedGroup } = useUrlState();
  const isSelected = selectedGroupId === groupId;
  const groupKey = sidebarNodeKey.group(groupId);
  const isOpen = useSidebarNavStore(
    (s) => !s.collapsedGroupIds.includes(groupId),
  );
  const isFocused = useSidebarNavStore((s) => s.focusedKey === groupKey);
  const focusedKey = useSidebarNavStore((s) => s.focusedKey);
  const setGroupCollapsed = useSidebarNavStore((s) => s.setGroupCollapsed);
  const setFocusedKey = useSidebarNavStore((s) => s.setFocusedKey);

  const isFocusInsideGroup =
    focusedKey !== null &&
    feeds.some((feed) => sidebarNodeKey.feed(feed.id) === focusedKey);

  // Collapsing hides the child feeds; keep the keyboard highlight on the group.
  const handleOpenChange = (open: boolean) => {
    setGroupCollapsed(groupId, !open);
    if (!open && isFocusInsideGroup) {
      setFocusedKey(groupKey);
    }
  };

  const unreadCount = feeds.reduce(
    (sum, feed) => sum + (feed.unread_count || 0),
    0,
  );

  return (
    <Collapsible
      open={isOpen}
      onOpenChange={handleOpenChange}
      className="w-full min-w-0"
    >
      <div
        data-sidebar-key={groupKey}
        className={cn(
          "flex w-full min-w-0 items-center gap-1.5 rounded-md px-2 py-1 text-sm transition-colors",
          isSelected
            ? "bg-accent text-accent-foreground"
            : "hover:bg-accent/50",
          isFocused && "ring-2 ring-inset ring-ring",
        )}
      >
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            handleOpenChange(!isOpen);
          }}
          className="shrink-0 p-1 -m-1 rounded-md transition-colors hover:bg-foreground/10"
        >
          <ChevronRight
            className={cn(
              "h-4 w-4 text-muted-foreground transition-transform",
              isOpen && "rotate-90",
            )}
          />
        </button>
        <button
          type="button"
          onClick={() => {
            setFocusedKey(groupKey);
            setSelectedGroup(groupId);
          }}
          className="flex min-w-0 flex-1 items-center gap-1.5 text-left"
        >
          <span className="block min-w-0 flex-1 truncate">{name}</span>
          {unreadCount > 0 && (
            <span className="shrink-0 text-[11px] text-muted-foreground">
              {unreadCount}
            </span>
          )}
        </button>
      </div>
      <CollapsibleContent>
        <div className="w-full min-w-0 pl-5">
          {feeds.map((feed) => (
            <FeedItem key={feed.id} feed={feed} />
          ))}
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
}
