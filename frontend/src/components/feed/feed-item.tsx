import { cn } from "@/lib/utils";
import { useUIStore } from "@/store";
import { getFaviconUrl } from "@/lib/api/favicon";
import { Settings } from "lucide-react";

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

  const handleSettingsClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    // TODO: Open feed settings modal
  };

  return (
    <button
      onClick={() => setSelectedFeed(id)}
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
      <span className="block min-w-0 max-w-full flex-1 truncate">{name}</span>
      <span className="ml-2 shrink-0 text-xs text-muted-foreground group-hover:hidden">
        {unreadCount > 0 ? unreadCount : ""}
      </span>
      <Settings
        className="ml-2 hidden h-4 w-4 shrink-0 text-muted-foreground hover:text-foreground group-hover:block"
        onClick={handleSettingsClick}
      />
    </button>
  );
}
