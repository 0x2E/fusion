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

interface FeedGroupProps {
  name: string;
  feeds: Feed[];
  unreadCount: number;
  getUnreadCount: (feedId: number) => number;
}

export function FeedGroup({
  name,
  feeds,
  unreadCount,
  getUnreadCount,
}: FeedGroupProps) {
  const [isOpen, setIsOpen] = useState(true);

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen}>
      <CollapsibleTrigger className="flex w-full items-center gap-2.5 rounded-md px-2 py-1.5 text-sm hover:bg-accent/50">
        <ChevronRight
          className={cn(
            "h-4 w-4 shrink-0 text-muted-foreground transition-transform",
            isOpen && "rotate-90",
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
              feedLink={feed.link}
              siteUrl={feed.site_url}
              unreadCount={getUnreadCount(feed.id)}
            />
          ))}
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
}
