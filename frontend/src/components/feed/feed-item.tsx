import { cn } from "@/lib/utils";
import { useUIStore } from "@/store";

interface FeedItemProps {
  id: number;
  name: string;
  unreadCount: number;
  color?: string;
}

export function FeedItem({ id, name, unreadCount, color = "#EB5757" }: FeedItemProps) {
  const { selectedFeedId, setSelectedFeed } = useUIStore();
  const isSelected = selectedFeedId === id;

  return (
    <button
      onClick={() => setSelectedFeed(id)}
      className={cn(
        "flex w-full items-center gap-2.5 rounded-md px-2 py-1.5 text-left text-sm transition-colors",
        isSelected
          ? "bg-accent text-accent-foreground"
          : "hover:bg-accent/50"
      )}
    >
      <div
        className="h-4 w-4 shrink-0 rounded"
        style={{ backgroundColor: color }}
      />
      <span className="flex-1 truncate">{name}</span>
      {unreadCount > 0 && (
        <span className="text-xs text-muted-foreground">{unreadCount}</span>
      )}
    </button>
  );
}
