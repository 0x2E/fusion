import { useRef, useState } from "react";
import { FileUp, Upload, X } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { useUIStore, useDataStore } from "@/store";
import { groupAPI, feedAPI } from "@/lib/api";
import { toast } from "sonner";
import { cn } from "@/lib/utils";
import { parseOPML } from "@/lib/opml";

export function ImportOpmlDialog() {
  const { isImportOpmlOpen, setImportOpmlOpen } = useUIStore();
  const { setGroups, setFeeds } = useDataStore();

  const [file, setFile] = useState<File | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [isImporting, setIsImporting] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const resetState = () => {
    setFile(null);
    setIsDragging(false);
  };

  const handleClose = () => {
    setImportOpmlOpen(false);
    resetState();
  };

  const handleFileSelect = (selectedFile: File | null) => {
    if (!selectedFile) return;

    const validTypes = [
      "text/xml",
      "application/xml",
      "text/x-opml",
      "application/x-opml",
    ];
    const validExtensions = [".opml", ".xml"];
    const hasValidExtension = validExtensions.some((ext) =>
      selectedFile.name.toLowerCase().endsWith(ext),
    );

    if (!validTypes.includes(selectedFile.type) && !hasValidExtension) {
      toast.error("Please select an OPML or XML file");
      return;
    }

    setFile(selectedFile);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const droppedFile = e.dataTransfer.files[0];
    handleFileSelect(droppedFile);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  };

  const handleImport = async () => {
    if (!file) return;

    setIsImporting(true);
    try {
      const content = await file.text();
      const parsedFeeds = parseOPML(content);

      if (parsedFeeds.length === 0) {
        toast.info("No feeds found in OPML file");
        return;
      }

      // Get existing groups
      const groupsRes = await groupAPI.list();
      const existingGroups = groupsRes.data;
      const groupNameToId = new Map(existingGroups.map((g) => [g.name, g.id]));

      // Collect unique group names that need to be created
      const newGroupNames = new Set<string>();
      parsedFeeds.forEach((feed) => {
        if (feed.groupName && !groupNameToId.has(feed.groupName)) {
          newGroupNames.add(feed.groupName);
        }
      });

      // Create missing groups
      for (const name of newGroupNames) {
        const res = await groupAPI.create({ name });
        if (res.data) {
          groupNameToId.set(name, res.data.id);
        }
      }

      // Use default group for feeds without a group
      let defaultGroupId: number | undefined = groupNameToId
        .values()
        .next().value;
      if (defaultGroupId === undefined) {
        const res = await groupAPI.create({ name: "Imported Without Group" });
        if (res.data) {
          defaultGroupId = res.data.id;
          groupNameToId.set("Imported Without Group", res.data.id);
        }
      }

      // Prepare batch create request
      const feedsToCreate = parsedFeeds.map((feed) => {
        const groupId = feed.groupName
          ? (groupNameToId.get(feed.groupName) ?? defaultGroupId!)
          : defaultGroupId!;
        return {
          group_id: groupId,
          name: feed.name,
          link: feed.link,
        };
      });

      const response = await feedAPI.batchCreate({ feeds: feedsToCreate });
      if (response.data) {
        const { created, failed, errors } = response.data;

        // Refresh groups and feeds after import
        const [newGroupsRes, feedsRes] = await Promise.all([
          groupAPI.list(),
          feedAPI.list(),
        ]);
        setGroups(newGroupsRes.data);
        setFeeds(feedsRes.data);

        if (created > 0) {
          toast.success(`Imported ${created} feed${created > 1 ? "s" : ""}`);
        }

        if (failed > 0) {
          const errorMsg = errors?.join(", ") || "Some feeds failed to import";
          toast.warning(
            `${failed} feed${failed > 1 ? "s" : ""} failed: ${errorMsg}`,
          );
        }

        if (created === 0 && failed === 0) {
          toast.info("No new feeds found in OPML file");
        }

        handleClose();
      }
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Failed to import OPML",
      );
    } finally {
      setIsImporting(false);
    }
  };

  return (
    <Dialog open={isImportOpmlOpen} onOpenChange={setImportOpmlOpen}>
      <DialogContent
        className="flex w-full max-w-[480px] flex-col gap-0 overflow-hidden p-0"
        showCloseButton={false}
      >
        <DialogHeader className="flex flex-row items-center justify-between border-b px-5 py-4">
          <DialogTitle className="text-base font-semibold">
            Import OPML
          </DialogTitle>
          <Button variant="ghost" size="icon-sm" onClick={handleClose}>
            <X className="h-[18px] w-[18px] text-muted-foreground" />
          </Button>
        </DialogHeader>

        <div className="p-5">
          <input
            ref={fileInputRef}
            type="file"
            accept=".opml,.xml,text/xml,application/xml"
            className="hidden"
            onChange={(e) => handleFileSelect(e.target.files?.[0] || null)}
          />

          <div
            onClick={() => fileInputRef.current?.click()}
            onDrop={handleDrop}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            className={cn(
              "flex cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed p-8 transition-colors",
              isDragging
                ? "border-primary bg-primary/5"
                : "border-muted-foreground/25 hover:border-muted-foreground/50",
              file && "border-primary bg-primary/5",
            )}
          >
            {file ? (
              <>
                <FileUp className="mb-3 h-10 w-10 text-primary" />
                <p className="text-sm font-medium">{file.name}</p>
                <p className="mt-1 text-xs text-muted-foreground">
                  {(file.size / 1024).toFixed(1)} KB
                </p>
                <Button
                  variant="link"
                  size="sm"
                  className="mt-2 h-auto p-0 text-xs"
                  onClick={(e) => {
                    e.stopPropagation();
                    setFile(null);
                  }}
                >
                  Choose a different file
                </Button>
              </>
            ) : (
              <>
                <Upload className="mb-3 h-10 w-10 text-muted-foreground" />
                <p className="text-sm font-medium">
                  Drop your OPML file here, or click to browse
                </p>
                <p className="mt-1 text-xs text-muted-foreground">
                  Supports .opml and .xml files
                </p>
              </>
            )}
          </div>
        </div>

        <div className="flex items-center justify-end gap-3 border-t px-5 py-4">
          <Button variant="outline" onClick={handleClose}>
            Cancel
          </Button>
          <Button onClick={handleImport} disabled={!file || isImporting}>
            <Upload className="mr-1.5 h-4 w-4" />
            {isImporting ? "Importing..." : "Import"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
