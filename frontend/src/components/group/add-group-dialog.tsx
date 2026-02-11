import { useState } from "react";
import { FolderPlus } from "lucide-react";
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
import { useUIStore } from "@/store";
import { useCreateGroup } from "@/queries/groups";
import { toast } from "sonner";

export function AddGroupDialog() {
  const { isAddGroupOpen, setAddGroupOpen } = useUIStore();
  const createGroup = useCreateGroup();

  const [name, setName] = useState("");

  const handleCreate = async () => {
    const trimmed = name.trim();
    if (!trimmed) return;

    try {
      await createGroup.mutateAsync(trimmed);
      setName("");
      setAddGroupOpen(false);
      toast.success("Group created");
    } catch {
      toast.error("Failed to create group");
    }
  };

  return (
    <Dialog
      open={isAddGroupOpen}
      onOpenChange={(open) => {
        setAddGroupOpen(open);
        if (!open) setName("");
      }}
    >
      <DialogContent className="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>Add Group</DialogTitle>
          <DialogDescription>
            Create a new group to organize your feeds.
          </DialogDescription>
        </DialogHeader>
        <div className="relative">
          <FolderPlus className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Group name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleCreate()}
            className="pl-10"
            autoFocus
          />
        </div>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => setAddGroupOpen(false)}
            disabled={createGroup.isPending}
          >
            Cancel
          </Button>
          <Button
            onClick={handleCreate}
            disabled={!name.trim() || createGroup.isPending}
          >
            {createGroup.isPending ? "Creating..." : "Create"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
