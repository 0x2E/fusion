import { useState, useEffect } from "react";
import { ChevronDown, Save, Trash2, X } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { Switch } from "@/components/ui/switch";
import { useUIStore } from "@/store";
import { useGroups } from "@/queries/groups";
import { useUpdateFeed, useDeleteFeed } from "@/queries/feeds";
import type { UpdateFeedRequest } from "@/lib/api";
import { toast } from "sonner";
import { cn } from "@/lib/utils";

export function EditFeedDialog() {
  const { isEditFeedOpen, editingFeed, setEditFeedOpen } = useUIStore();
  const { data: groups = [] } = useGroups();
  const updateFeedMutation = useUpdateFeed();
  const deleteFeedMutation = useDeleteFeed();

  const [url, setUrl] = useState("");
  const [name, setName] = useState("");
  const [groupId, setGroupId] = useState<string>("");
  const [proxy, setProxy] = useState("");
  const [suspended, setSuspended] = useState(false);
  const [isAdvancedOpen, setIsAdvancedOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  useEffect(() => {
    if (editingFeed) {
      setUrl(editingFeed.link);
      setName(editingFeed.name);
      setGroupId(editingFeed.group_id.toString());
      setProxy(editingFeed.proxy ?? "");
      setSuspended(editingFeed.suspended);
      setIsAdvancedOpen(!!editingFeed.proxy);
    }
  }, [editingFeed]);

  const resetForm = () => {
    setUrl("");
    setName("");
    setGroupId("");
    setProxy("");
    setSuspended(false);
    setIsAdvancedOpen(false);
    setIsDeleteOpen(false);
  };

  const handleClose = () => {
    setEditFeedOpen(false);
    resetForm();
  };

  const handleSubmit = async () => {
    if (!editingFeed) return;

    if (!url.trim()) {
      toast.error("Please enter a feed URL");
      return;
    }

    if (!name.trim()) {
      toast.error("Please enter a feed name");
      return;
    }

    setIsSubmitting(true);
    try {
      const request: UpdateFeedRequest = {};

      if (url.trim() !== editingFeed.link) {
        request.link = url.trim();
      }

      if (name.trim() !== editingFeed.name) {
        request.name = name.trim();
      }

      const newGroupId = parseInt(groupId, 10);
      if (newGroupId !== editingFeed.group_id) {
        request.group_id = newGroupId;
      }

      if (suspended !== editingFeed.suspended) {
        request.suspended = suspended;
      }

      const newProxy = proxy.trim() || undefined;
      if (newProxy !== editingFeed.proxy) {
        request.proxy = newProxy;
      }

      if (Object.keys(request).length === 0) {
        toast.info("No changes to save");
        handleClose();
        return;
      }

      await updateFeedMutation.mutateAsync({ id: editingFeed.id, ...request });
      toast.success("Feed updated successfully");
      handleClose();
    } catch {
      toast.error("Failed to update feed");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!editingFeed) return;

    setIsDeleting(true);
    try {
      await deleteFeedMutation.mutateAsync(editingFeed.id);
      toast.success(`Unsubscribed from "${editingFeed.name}"`);
      setIsDeleteOpen(false);
      handleClose();
    } catch {
      toast.error("Failed to unsubscribe from feed");
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <>
      <Dialog
        open={isEditFeedOpen}
        onOpenChange={(open) => setEditFeedOpen(open)}
      >
        <DialogContent
          className="flex w-full max-w-[480px] flex-col gap-0 overflow-hidden p-0"
          showCloseButton={false}
        >
          {/* Header */}
          <DialogHeader className="flex flex-row items-center justify-between border-b px-5 py-4">
            <DialogTitle className="text-base font-semibold">
              Edit Feed
            </DialogTitle>
            <Button variant="ghost" size="icon-sm" onClick={handleClose}>
              <X className="h-[18px] w-[18px] text-muted-foreground" />
            </Button>
          </DialogHeader>

          {/* Form Content */}
          <div className="space-y-4 p-5">
            {/* URL Section */}
            <div className="space-y-1.5">
              <label className="text-[13px] font-medium">Feed URL</label>
              <Input
                placeholder="https://example.com/feed.xml"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                className="h-10"
              />
            </div>

            {/* Name Section */}
            <div className="space-y-1.5">
              <label className="text-[13px] font-medium">Feed Name</label>
              <Input
                placeholder="Enter feed name..."
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="h-10"
              />
            </div>

            {/* Group Section */}
            <div className="space-y-1.5">
              <label className="text-[13px] font-medium">Group</label>
              <Select value={groupId} onValueChange={setGroupId}>
                <SelectTrigger className="h-10">
                  <SelectValue placeholder="Select a group..." />
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

            {/* Suspended Toggle */}
            <div className="flex items-center justify-between">
              <div>
                <label className="text-[13px] font-medium">Suspend Feed</label>
                <p className="text-xs text-muted-foreground">
                  Pauses automatic updates
                </p>
              </div>
              <Switch checked={suspended} onCheckedChange={setSuspended} />
            </div>

            {/* Advanced Section */}
            <Collapsible open={isAdvancedOpen} onOpenChange={setIsAdvancedOpen}>
              <CollapsibleTrigger className="flex w-full items-center gap-1.5 text-[13px] font-medium text-muted-foreground">
                <ChevronDown
                  className={cn(
                    "h-3.5 w-3.5 transition-transform",
                    isAdvancedOpen && "rotate-180",
                  )}
                />
                Advanced Settings
              </CollapsibleTrigger>
              <CollapsibleContent className="space-y-1.5 pl-5 pt-3">
                <label className="text-[13px] font-medium">HTTP Proxy</label>
                <Input
                  placeholder="http://proxy.example.com:8080"
                  value={proxy}
                  onChange={(e) => setProxy(e.target.value)}
                  className="h-10"
                />
                <p className="text-xs text-muted-foreground">
                  Leave empty to use system proxy settings
                </p>
              </CollapsibleContent>
            </Collapsible>
          </div>

          {/* Footer */}
          <div className="flex items-center justify-between border-t px-5 py-4">
            <Button
              variant="ghost"
              size="sm"
              className="text-destructive hover:text-destructive hover:bg-destructive/10"
              onClick={() => setIsDeleteOpen(true)}
            >
              <Trash2 className="mr-1.5 h-3.5 w-3.5" />
              Unsubscribe
            </Button>
            <div className="flex items-center gap-3">
              <Button variant="outline" onClick={handleClose}>
                Cancel
              </Button>
              <Button
                onClick={handleSubmit}
                disabled={isSubmitting || !url.trim() || !name.trim()}
              >
                <Save className="mr-1.5 h-4 w-4" />
                Save
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={isDeleteOpen} onOpenChange={setIsDeleteOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Unsubscribe Feed</DialogTitle>
            <DialogDescription>
              Are you sure you want to unsubscribe from{" "}
              <span className="font-semibold">{editingFeed?.name}</span>? All
              items from this feed will be deleted.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsDeleteOpen(false)}
              disabled={isDeleting}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={isDeleting}
            >
              {isDeleting ? "Deleting..." : "Unsubscribe"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
