import { cn } from "@/lib/utils";
import { useUrlState } from "@/hooks/use-url-state";
import { useUIStore } from "@/store";
import { getFaviconUrl } from "@/lib/api/favicon";
import type { Feed } from "@/lib/api";
import { FeedFavicon } from "@/components/feed/feed-favicon";
import { Settings } from "lucide-react";
import { Button } from "@/components/ui/button";

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
        "group flex w-full min-w-0 items-center gap-2 rounded-md px-2 py-1 text-left text-sm transition-colors",
        isSelected ? "bg-accent text-accent-foreground" : "hover:bg-accent/50",
      )}
    >
      <FeedFavicon src={faviconUrl} className="h-4 w-4" />
      <span className="block min-w-0 max-w-full flex-1 truncate">
        {feed.name}
      </span>
      <div className="ml-2 flex h-6 shrink-0 items-center justify-center">
        <span className="text-xs text-muted-foreground/60 group-hover:hidden">
          {feed.unread_count > 0 ? feed.unread_count : ""}
        </span>
        <Button
          variant="ghost"
          size="icon-xs"
          className="hidden group-hover:inline-flex"
          onClick={handleSettingsClick}
        >
          <Settings className="text-muted-foreground" />
        </Button>
      </div>
    </button>
  );
}
