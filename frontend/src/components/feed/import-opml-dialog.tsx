import { useRef, useState } from "react";
import {
  CircleCheck,
  FileUp,
  OctagonX,
  TriangleAlert,
  Upload,
  X,
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { useUIStore } from "@/store";
import { useQueryClient } from "@tanstack/react-query";
import { groupAPI, feedAPI } from "@/lib/api";
import { queryKeys } from "@/queries/keys";
import { toast } from "sonner";
import { useI18n } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import { parseOPML } from "@/lib/opml";

export function ImportOpmlDialog() {
  const { t } = useI18n();
  const { isImportOpmlOpen, setImportOpmlOpen } = useUIStore();
  const queryClient = useQueryClient();
  const fileInputId = "opml-file-input";

  const [file, setFile] = useState<File | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [isImporting, setIsImporting] = useState(false);
  const [importResult, setImportResult] = useState<{
    created: number;
    failed: number;
    errors: string[];
  } | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const resetState = () => {
    setFile(null);
    setIsDragging(false);
    setImportResult(null);
  };

  const handleOpenChange = (open: boolean) => {
    // Closing mid-import would let the late response repopulate the result
    // view, so a reopened dialog would show a stale import.
    if (!open && isImporting) return;
    setImportOpmlOpen(open);
    if (!open) resetState();
  };

  const handleClose = () => handleOpenChange(false);

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
      toast.error(t("opml.toast.invalidFile"));
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
        toast.info(t("opml.toast.noFeeds"));
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
        const res = await groupAPI.create({ name: t("opml.group.default") });
        if (res.data) {
          defaultGroupId = res.data.id;
          groupNameToId.set(t("opml.group.default"), res.data.id);
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
          ...(feed.siteUrl ? { site_url: feed.siteUrl } : {}),
        };
      });

      const response = await feedAPI.batchCreate({ feeds: feedsToCreate });
      if (response.data) {
        const { created, failed, errors } = response.data;

        // Invalidate all caches after import
        await Promise.all([
          queryClient.invalidateQueries({ queryKey: queryKeys.groups.all }),
          queryClient.invalidateQueries({ queryKey: queryKeys.feeds.all }),
          queryClient.invalidateQueries({ queryKey: queryKeys.items.all }),
        ]);

        setImportResult({ created, failed, errors: errors ?? [] });
      }
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : t("opml.toast.failed"),
      );
    } finally {
      setIsImporting(false);
    }
  };

  return (
    <Dialog open={isImportOpmlOpen} onOpenChange={handleOpenChange}>
      <DialogContent
        className="flex w-full max-w-[480px] flex-col gap-0 overflow-hidden p-0"
        showCloseButton={false}
      >
        <DialogHeader className="flex flex-row items-center justify-between border-b px-5 py-4">
          <DialogTitle className="text-base font-semibold">
            {t("opml.dialog.title")}
          </DialogTitle>
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={handleClose}
            disabled={isImporting}
          >
            <span className="sr-only">{t("common.cancel")}</span>
            <X className="h-[18px] w-[18px] text-muted-foreground" />
          </Button>
        </DialogHeader>

        <div className="p-5">
          <input
            ref={fileInputRef}
            id={fileInputId}
            type="file"
            accept=".opml,.xml,text/xml,application/xml"
            className="hidden"
            aria-label={t("opml.dialog.title")}
            onChange={(e) => handleFileSelect(e.target.files?.[0] || null)}
          />

          {importResult ? (
            <div className="flex flex-col gap-3">
              <div role="status" className="flex gap-3">
                {importResult.failed === 0 ? (
                  <CircleCheck className="mt-0.5 size-5 shrink-0 text-primary" />
                ) : importResult.created > 0 ? (
                  <TriangleAlert className="mt-0.5 size-5 shrink-0 text-amber-500" />
                ) : (
                  <OctagonX className="mt-0.5 size-5 shrink-0 text-destructive" />
                )}
                <div>
                  <p className="text-sm font-medium">
                    {t("opml.result.title")}
                  </p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {importResult.failed > 0
                      ? t("opml.result.summary", {
                          created: importResult.created,
                          failed: importResult.failed,
                        })
                      : t("opml.result.imported", {
                          count: importResult.created,
                        })}
                  </p>
                </div>
              </div>
              {importResult.errors.length > 0 && (
                <ul className="max-h-64 divide-y overflow-y-auto overscroll-contain rounded-md border text-xs">
                  {importResult.errors.map((error, index) => (
                    <li
                      key={`${index}-${error}`}
                      className="px-3 py-2 break-words text-muted-foreground"
                    >
                      {error}
                    </li>
                  ))}
                </ul>
              )}
            </div>
          ) : (
            <label
              htmlFor={fileInputId}
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
                      e.preventDefault();
                      e.stopPropagation();
                      setFile(null);
                    }}
                  >
                    {t("opml.dialog.chooseDifferentFile")}
                  </Button>
                </>
              ) : (
                <>
                  <Upload className="mb-3 h-10 w-10 text-muted-foreground" />
                  <p className="text-sm font-medium">
                    {t("opml.dialog.dropHint")}
                  </p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {t("opml.dialog.supportHint")}
                  </p>
                </>
              )}
            </label>
          )}
        </div>

        <div className="flex items-center justify-end gap-3 border-t px-5 py-4">
          {importResult ? (
            <Button onClick={handleClose}>{t("opml.result.done")}</Button>
          ) : (
            <>
              <Button
                variant="outline"
                onClick={handleClose}
                disabled={isImporting}
              >
                {t("common.cancel")}
              </Button>
              <Button onClick={handleImport} disabled={!file || isImporting}>
                <Upload className="mr-1.5 h-4 w-4" />
                {isImporting ? t("common.importing") : t("common.import")}
              </Button>
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
