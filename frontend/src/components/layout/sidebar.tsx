import { Search, Settings } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { FeedList } from "@/components/feed/feed-list";
import { useUIStore } from "@/store";

export function Sidebar() {
  const { setSearchOpen, setSettingsOpen } = useUIStore();

  return (
    <aside className="flex h-full w-[260px] shrink-0 flex-col border-r bg-sidebar">
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-3">
        <div className="flex h-7 w-7 items-center justify-center rounded-md bg-primary">
          <span className="text-sm font-bold text-primary-foreground">F</span>
        </div>
        <span className="text-base font-semibold">Fusion</span>
      </div>

      {/* Search button */}
      <div className="px-2">
        <Button
          variant="ghost"
          className="w-full justify-start gap-2 text-muted-foreground"
          onClick={() => setSearchOpen(true)}
        >
          <Search className="h-4 w-4" />
          <span className="flex-1 text-left text-sm">Search</span>
          <kbd className="pointer-events-none hidden h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium opacity-100 sm:flex">
            <span className="text-xs">⌘</span>K
          </kbd>
        </Button>
      </div>

      <Separator className="my-2" />

      {/* Feed list */}
      <FeedList />

      {/* Footer */}
      <Separator />
      <div className="p-2">
        <Button
          variant="ghost"
          className="w-full justify-start gap-2 text-muted-foreground"
          onClick={() => setSettingsOpen(true)}
        >
          <Settings className="h-4 w-4" />
          <span className="text-sm">Settings</span>
        </Button>
      </div>
    </aside>
  );
}
