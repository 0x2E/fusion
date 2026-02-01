import { useState, useMemo } from "react";
import { Download, Pencil, Search, Trash2, X } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useUIStore, useDataStore } from "@/store";
import { feedAPI, groupAPI, type Feed } from "@/lib/api";
import { toast } from "sonner";
import { generateOPML, downloadFile } from "@/lib/opml";

export function FeedManagementDialog() {
  const { isFeedManagementOpen, setFeedManagementOpen } = useUIStore();
  const { feeds, removeFeed, getGroupById } = useDataStore();

  const [searchQuery, setSearchQuery] = useState("");
  const [deletingFeedId, setDeletingFeedId] = useState<number | null>(null);

  const filteredFeeds = useMemo(() => {
    if (!searchQuery.trim()) return feeds;
    const query = searchQuery.toLowerCase();
    return feeds.filter(
      (feed) =>
        feed.name.toLowerCase().includes(query) ||
        feed.link.toLowerCase().includes(query),
    );
  }, [feeds, searchQuery]);

  const handleClose = () => {
    setFeedManagementOpen(false);
    setSearchQuery("");
  };

  const [isExporting, setIsExporting] = useState(false);

  const handleExport = async () => {
    setIsExporting(true);
    try {
      const [groupsRes, feedsRes] = await Promise.all([
        groupAPI.list(),
        feedAPI.list(),
      ]);

      const opml = generateOPML(groupsRes.data, feedsRes.data);
      downloadFile(opml, "fusion-subscriptions.opml", "application/xml");
      toast.success("OPML exported successfully");
    } catch {
      toast.error("Failed to export OPML");
    } finally {
      setIsExporting(false);
    }
  };

  const handleDelete = async (feed: Feed) => {
    setDeletingFeedId(feed.id);
    try {
      await feedAPI.delete(feed.id);
      removeFeed(feed.id);
      toast.success(`Unsubscribed from "${feed.name}"`);
    } catch {
      toast.error("Failed to unsubscribe from feed");
    } finally {
      setDeletingFeedId(null);
    }
  };

  const getGroupName = (groupId: number) => {
    const group = getGroupById(groupId);
    return group?.name ?? "Ungrouped";
  };

  const getDomain = (url: string) => {
    try {
      return new URL(url).hostname;
    } catch {
      return url;
    }
  };

  return (
    <Dialog open={isFeedManagementOpen} onOpenChange={setFeedManagementOpen}>
      <DialogContent
        className="flex max-h-[85vh] w-[560px] flex-col gap-0 overflow-hidden p-0"
        showCloseButton={false}
      >
        {/* Header */}
        <DialogHeader className="flex flex-row items-center justify-between border-b px-5 py-4">
          <DialogTitle className="text-base font-semibold">
            Manage Feeds
          </DialogTitle>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              className="gap-1.5"
              onClick={handleExport}
              disabled={isExporting}
            >
              <Download className="h-3.5 w-3.5" />
              {isExporting ? "Exporting..." : "Export OPML"}
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={handleClose}
            >
              <X className="h-[18px] w-[18px] text-muted-foreground" />
            </Button>
          </div>
        </DialogHeader>

        {/* Search Section */}
        <div className="border-b px-5 py-3">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search feeds..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="h-9 pl-9"
            />
          </div>
        </div>

        {/* Feed List */}
        <div className="flex-1 overflow-hidden px-3 py-2">
          <span className="px-2 text-[11px] font-semibold uppercase text-muted-foreground">
            All Feeds ({filteredFeeds.length})
          </span>
          <ScrollArea className="mt-2 h-[400px]">
            <div className="space-y-0.5">
              {filteredFeeds.map((feed) => (
                <div
                  key={feed.id}
                  className="flex items-center justify-between rounded-md px-2 py-2.5 hover:bg-accent/50"
                >
                  <div className="flex min-w-0 items-center gap-2.5">
                    <div
                      className="h-[18px] w-[18px] shrink-0 rounded"
                      style={{ backgroundColor: getFeedColor(feed.id) }}
                    />
                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium">
                        {feed.name}
                      </p>
                      <p className="truncate text-xs text-muted-foreground">
                        {getDomain(feed.link)}
                      </p>
                    </div>
                  </div>
                  <div className="flex shrink-0 items-center gap-2">
                    <span className="rounded bg-accent px-2 py-0.5 text-[11px] font-medium text-muted-foreground">
                      {getGroupName(feed.group_id)}
                    </span>
                    <div className="flex gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7"
                        onClick={() => {
                          // TODO: Open feed settings modal
                          toast.info("Feed settings coming soon");
                        }}
                      >
                        <Pencil className="h-3.5 w-3.5 text-muted-foreground" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7"
                        onClick={() => handleDelete(feed)}
                        disabled={deletingFeedId === feed.id}
                      >
                        <Trash2 className="h-3.5 w-3.5 text-muted-foreground" />
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
              {filteredFeeds.length === 0 && (
                <div className="py-8 text-center text-sm text-muted-foreground">
                  {searchQuery ? "No feeds match your search" : "No feeds yet"}
                </div>
              )}
            </div>
          </ScrollArea>
        </div>
      </DialogContent>
    </Dialog>
  );
}

// Generate a consistent color based on feed ID
function getFeedColor(id: number): string {
  const colors = [
    "#FF6600",
    "#1DA1F2",
    "#5865F2",
    "#EB5757",
    "#2D9CDB",
    "#27AE60",
    "#9B51E0",
    "#F2994A",
  ];
  return colors[id % colors.length];
}
