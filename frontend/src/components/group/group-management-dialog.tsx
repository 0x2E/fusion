import { useState } from "react";
import { Folder, FolderPlus, Pencil, Trash2, X } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useUIStore, useDataStore } from "@/store";
import { groupAPI, feedAPI, type Group } from "@/lib/api";
import { toast } from "sonner";

export function GroupManagementDialog() {
  const { isGroupManagementOpen, setGroupManagementOpen } = useUIStore();
  const {
    groups,
    feeds,
    addGroup,
    updateGroup,
    removeGroup,
    moveFeedsToGroup,
  } = useDataStore();

  const [newGroupName, setNewGroupName] = useState("");
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editingName, setEditingName] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  const [deletingGroup, setDeletingGroup] = useState<Group | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const getFeedCount = (groupId: number) =>
    feeds.filter((f) => f.group_id === groupId).length;

  const handleCreate = async () => {
    const name = newGroupName.trim();
    if (!name) return;

    setIsCreating(true);
    try {
      const response = await groupAPI.create({ name });
      if (response.data) {
        addGroup(response.data);
        setNewGroupName("");
        toast.success("Group created");
      }
    } catch {
      toast.error("Failed to create group");
    } finally {
      setIsCreating(false);
    }
  };

  const handleUpdate = async (group: Group) => {
    const name = editingName.trim();
    if (!name || name === group.name) {
      setEditingId(null);
      return;
    }

    try {
      await groupAPI.update(group.id, { name });
      updateGroup(group.id, name);
      setEditingId(null);
      toast.success("Group updated");
    } catch {
      toast.error("Failed to update group");
    }
  };

  const confirmDelete = async () => {
    if (!deletingGroup) return;

    setIsDeleting(true);
    try {
      const groupFeeds = feeds.filter((f) => f.group_id === deletingGroup.id);

      // Move all feeds to group id=1
      await Promise.all(
        groupFeeds.map((feed) => feedAPI.update(feed.id, { group_id: 1 })),
      );

      // Delete the group
      await groupAPI.delete(deletingGroup.id);

      // Update local state
      moveFeedsToGroup(deletingGroup.id, 1);
      removeGroup(deletingGroup.id);

      toast.success("Group deleted");
      setDeletingGroup(null);
    } catch {
      toast.error("Failed to delete group");
    } finally {
      setIsDeleting(false);
    }
  };

  const startEditing = (group: Group) => {
    setEditingId(group.id);
    setEditingName(group.name);
  };

  const targetGroup = groups.find((g) => g.id === 1);

  return (
    <>
      <Dialog
        open={isGroupManagementOpen}
        onOpenChange={setGroupManagementOpen}
      >
        <DialogContent
          className="flex max-h-[85vh] w-[480px] flex-col gap-0 overflow-hidden p-0"
          showCloseButton={false}
        >
          {/* Header */}
          <DialogHeader className="flex flex-row items-center justify-between border-b px-5 py-4">
            <DialogTitle className="text-base font-semibold">
              Group Management
            </DialogTitle>
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={() => setGroupManagementOpen(false)}
            >
              <X className="h-[18px] w-[18px] text-muted-foreground" />
            </Button>
          </DialogHeader>

          {/* Add Group Section */}
          <div className="flex items-center gap-3 border-b px-5 py-3">
            <div className="relative flex-1">
              <FolderPlus className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Enter new group name..."
                value={newGroupName}
                onChange={(e) => setNewGroupName(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleCreate()}
                className="h-10 pl-10"
              />
            </div>
            <Button
              onClick={handleCreate}
              disabled={!newGroupName.trim() || isCreating}
              className="h-10 gap-1.5"
            >
              <span>Add</span>
            </Button>
          </div>

          {/* Group List */}
          <div className="flex-1 overflow-hidden">
            <div className="px-3 py-2">
              <span className="text-[11px] font-semibold uppercase text-muted-foreground">
                Your Groups
              </span>
            </div>
            <ScrollArea className="h-[300px] px-3 pb-3">
              <div className="space-y-1">
                {groups.map((group) => {
                  const feedCount = getFeedCount(group.id);
                  const isEditing = editingId === group.id;

                  return (
                    <div
                      key={group.id}
                      className="flex items-center justify-between rounded-md bg-accent/50 px-2 py-2.5"
                    >
                      <div className="flex items-center gap-2.5">
                        <Folder className="h-[18px] w-[18px] text-foreground" />
                        {isEditing ? (
                          <Input
                            value={editingName}
                            onChange={(e) => setEditingName(e.target.value)}
                            onBlur={() => handleUpdate(group)}
                            onKeyDown={(e) => {
                              if (e.key === "Enter") handleUpdate(group);
                              if (e.key === "Escape") setEditingId(null);
                            }}
                            className="h-7 w-40 px-2 text-sm"
                            autoFocus
                          />
                        ) : (
                          <>
                            <span className="text-sm font-medium">
                              {group.name}
                            </span>
                            <span className="text-xs text-muted-foreground">
                              {feedCount} {feedCount === 1 ? "feed" : "feeds"}
                            </span>
                          </>
                        )}
                      </div>
                      <div className="flex items-center gap-1">
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          onClick={() => startEditing(group)}
                        >
                          <Pencil className="h-3.5 w-3.5 text-muted-foreground" />
                        </Button>
                        {group.id !== 1 && (
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            onClick={() => setDeletingGroup(group)}
                          >
                            <Trash2 className="h-3.5 w-3.5 text-muted-foreground" />
                          </Button>
                        )}
                      </div>
                    </div>
                  );
                })}
                {groups.length === 0 && (
                  <div className="py-8 text-center text-sm text-muted-foreground">
                    No groups yet. Create one above.
                  </div>
                )}
              </div>
            </ScrollArea>
          </div>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog
        open={deletingGroup !== null}
        onOpenChange={(open) => !open && setDeletingGroup(null)}
      >
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Delete Group</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete{" "}
              <span className="font-semibold">{deletingGroup?.name}</span>?
              {getFeedCount(deletingGroup?.id ?? 0) > 0 ? (
                <>
                  {" "}
                  All {getFeedCount(deletingGroup?.id ?? 0)} feed(s) in this
                  group will be moved to{" "}
                  <span className="font-semibold">
                    {targetGroup?.name ?? "Default"}
                  </span>
                  .
                </>
              ) : (
                ""
              )}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeletingGroup(null)}
              disabled={isDeleting}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={confirmDelete}
              disabled={isDeleting}
            >
              {isDeleting ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
