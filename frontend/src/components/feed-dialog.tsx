import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { feedAPI } from "@/lib/api";
import type { Feed } from "@/lib/api/types";
import { useDataStore } from "@/store/data";
import { useEffect, useState } from "react";
import { toast } from "sonner";

interface FeedDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  feed?: Feed;
  defaultGroupId?: number;
}

export function FeedDialog({
  open,
  onOpenChange,
  feed,
  defaultGroupId,
}: FeedDialogProps) {
  const [formData, setFormData] = useState({
    groupId: defaultGroupId || 0,
    name: "",
    url: "",
    siteUrl: "",
    proxy: "",
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isValidating, setIsValidating] = useState(false);
  const { groups, createFeed, updateFeed } = useDataStore();

  useEffect(() => {
    if (open) {
      if (feed) {
        setFormData({
          groupId: feed.group_id,
          name: feed.name,
          url: feed.link,
          siteUrl: feed.site_url || "",
          proxy: feed.proxy || "",
        });
      } else {
        setFormData({
          groupId: defaultGroupId || groups[0]?.id || 0,
          name: "",
          url: "",
          siteUrl: "",
          proxy: "",
        });
      }
    }
  }, [open, feed, defaultGroupId, groups]);

  const handleValidate = async () => {
    if (!formData.url.trim()) {
      toast.error("Please enter a feed URL");
      return;
    }

    setIsValidating(true);
    try {
      const response = await feedAPI.validate({ url: formData.url });
      if (response.data?.valid) {
        toast.success("Feed URL is valid");
      } else {
        toast.error("Feed URL is invalid");
      }
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Failed to validate feed"
      );
    } finally {
      setIsValidating(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const trimmedUrl = formData.url.trim();
    if (!trimmedUrl || !formData.groupId) {
      return;
    }

    setIsSubmitting(true);
    try {
      if (feed) {
        await updateFeed(feed.id, {
          group_id: formData.groupId,
          name: formData.name.trim() || undefined,
          site_url: formData.siteUrl.trim() || undefined,
          proxy: formData.proxy.trim() || undefined,
        });
      } else {
        await createFeed({
          group_id: formData.groupId,
          name: formData.name.trim(),
          link: trimmedUrl,
          site_url: formData.siteUrl.trim() || undefined,
          proxy: formData.proxy.trim() || undefined,
        });
      }
      onOpenChange(false);
      setFormData({
        groupId: 0,
        name: "",
        url: "",
        siteUrl: "",
        proxy: "",
      });
    } catch (error) {
      // Error is already handled in store with toast
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{feed ? "Edit Feed" : "Add Feed"}</DialogTitle>
          </DialogHeader>

          <div className="py-4 space-y-4">
            <div className="space-y-2">
              <Label htmlFor="feed-url">
                Feed URL <span className="text-destructive">*</span>
              </Label>
              <div className="flex gap-2">
                <Input
                  id="feed-url"
                  placeholder="https://example.com/feed.xml"
                  value={formData.url}
                  onChange={(e) =>
                    setFormData({ ...formData, url: e.target.value })
                  }
                  autoFocus
                  required
                  disabled={!!feed}
                  className="flex-1"
                />
                {!feed && (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleValidate}
                    disabled={isValidating || !formData.url.trim()}
                  >
                    {isValidating ? "Checking..." : "Validate"}
                  </Button>
                )}
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="feed-name">Feed Name</Label>
              <Input
                id="feed-name"
                placeholder="Leave empty to use feed title"
                value={formData.name}
                onChange={(e) =>
                  setFormData({ ...formData, name: e.target.value })
                }
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="feed-group">
                Group <span className="text-destructive">*</span>
              </Label>
              <Select
                value={formData.groupId.toString()}
                onValueChange={(value) =>
                  setFormData({ ...formData, groupId: parseInt(value) })
                }
                required
              >
                <SelectTrigger id="feed-group">
                  <SelectValue placeholder="Select a group" />
                </SelectTrigger>
                <SelectContent>
                  {groups.map((group) => (
                    <SelectItem key={group.id} value={group.id.toString()}>
                      {group.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="feed-site-url">Site URL</Label>
              <Input
                id="feed-site-url"
                placeholder="https://example.com"
                value={formData.siteUrl}
                onChange={(e) =>
                  setFormData({ ...formData, siteUrl: e.target.value })
                }
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="feed-proxy">Proxy</Label>
              <Input
                id="feed-proxy"
                placeholder="http://proxy.example.com:8080"
                value={formData.proxy}
                onChange={(e) =>
                  setFormData({ ...formData, proxy: e.target.value })
                }
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="ghost"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={
                isSubmitting || !formData.url.trim() || !formData.groupId
              }
            >
              {isSubmitting ? "Saving..." : feed ? "Save" : "Add"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
