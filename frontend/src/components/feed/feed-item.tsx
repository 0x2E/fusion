import { cn } from "@/lib/utils";
import { useUIStore } from "@/store";
import { getFaviconUrl } from "@/lib/api/favicon";

interface FeedItemProps {
  id: number;
  name: string;
  feedLink: string;
  siteUrl?: string;
  unreadCount: number;
}

export function FeedItem({
  id,
  name,
  feedLink,
  siteUrl,
  unreadCount,
}: FeedItemProps) {
  const { selectedFeedId, setSelectedFeed } = useUIStore();
  const isSelected = selectedFeedId === id;
  const faviconUrl = getFaviconUrl(feedLink, siteUrl);

  return (
    <button
      onClick={() => setSelectedFeed(id)}
      className={cn(
        "flex w-full items-center gap-2.5 rounded-md px-2 py-1.5 text-left text-sm transition-colors",
        isSelected ? "bg-accent text-accent-foreground" : "hover:bg-accent/50",
      )}
    >
      <img
        src={faviconUrl}
        alt=""
        className="h-4 w-4 shrink-0 rounded"
        loading="lazy"
      />
      <span className="flex-1 truncate">{name}</span>
      {unreadCount > 0 && (
        <span className="text-xs text-muted-foreground">{unreadCount}</span>
      )}
    </button>
  );
}
