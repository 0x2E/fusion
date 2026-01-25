import { useState } from "react";
import { ChevronRight } from "lucide-react";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { cn } from "@/lib/utils";
import { FeedItem } from "./feed-item";
import type { Feed } from "@/lib/api";

// Generate consistent color from feed name
function getColorFromName(name: string): string {
  const colors = [
    "#EB5757", "#F2994A", "#F2C94C", "#27AE60",
    "#2D9CDB", "#9B51E0", "#BB6BD9", "#56CCF2",
  ];
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  return colors[Math.abs(hash) % colors.length];
}

interface FeedGroupProps {
  name: string;
  feeds: Feed[];
  unreadCount: number;
  getUnreadCount: (feedId: number) => number;
}

export function FeedGroup({ name, feeds, unreadCount, getUnreadCount }: FeedGroupProps) {
  const [isOpen, setIsOpen] = useState(true);

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen}>
      <CollapsibleTrigger className="flex w-full items-center gap-1 rounded-md px-2 py-1.5 text-sm font-medium hover:bg-accent/50">
        <ChevronRight
          className={cn(
            "h-4 w-4 shrink-0 text-muted-foreground transition-transform",
            isOpen && "rotate-90"
          )}
        />
        <span className="flex-1 truncate text-left">{name}</span>
        {unreadCount > 0 && (
          <span className="text-xs text-muted-foreground">{unreadCount}</span>
        )}
      </CollapsibleTrigger>
      <CollapsibleContent>
        <div className="ml-3 mt-0.5 space-y-0.5 border-l pl-2">
          {feeds.map((feed) => (
            <FeedItem
              key={feed.id}
              id={feed.id}
              name={feed.name}
              unreadCount={getUnreadCount(feed.id)}
              color={getColorFromName(feed.name)}
            />
          ))}
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
}
