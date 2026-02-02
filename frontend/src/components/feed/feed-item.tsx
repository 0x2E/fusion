import { cn } from "@/lib/utils";
import { useUrlState } from "@/hooks/use-url-state";
import { useUIStore } from "@/store";
import { getFaviconUrl } from "@/lib/api/favicon";
import type { Feed } from "@/lib/api";
import { Settings } from "lucide-react";

interface FeedItemProps {
  feed: Feed;
}

export function FeedItem({ feed }: FeedItemProps) {
  const { selectedFeedId, setSelectedFeed } = useUrlState();
  const { setEditFeedOpen } = useUIStore();

  const isSelected = selectedFeedId === feed.id;
  const faviconUrl = getFaviconUrl(feed.link, feed.site_url);

  const handleSettingsClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    setEditFeedOpen(true, feed);
  };

  return (
    <button
      onClick={() => setSelectedFeed(feed.id)}
      className={cn(
        "group flex w-full min-w-0 items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors",
        isSelected ? "bg-accent text-accent-foreground" : "hover:bg-accent/50",
      )}
    >
      <img
        src={faviconUrl}
        alt=""
        className="h-4 w-4 shrink-0 rounded"
        loading="lazy"
      />
      <span className="block min-w-0 max-w-full flex-1 truncate">
        {feed.name}
      </span>
      <span className="ml-2 shrink-0 text-xs text-muted-foreground group-hover:hidden">
        {feed.unread_count > 0 ? feed.unread_count : ""}
      </span>
      <Settings
        className="ml-2 hidden h-4 w-4 shrink-0 text-muted-foreground hover:text-foreground group-hover:block"
        onClick={handleSettingsClick}
      />
    </button>
  );
}
