import { useState } from "react";
import { ChevronDown, Plus, Radar, X } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
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
import { useUIStore } from "@/store";
import { useGroups } from "@/queries/groups";
import { useCreateFeed } from "@/queries/feeds";
import {
  feedAPI,
  type CreateFeedRequest,
  type DiscoveredFeed,
} from "@/lib/api";
import { toast } from "sonner";
import { useI18n } from "@/lib/i18n";
import { cn } from "@/lib/utils";

export function AddFeedDialog() {
  const { t } = useI18n();
  const { isAddFeedOpen, setAddFeedOpen } = useUIStore();
  const { data: groups = [] } = useGroups();
  const createFeed = useCreateFeed();

  const [url, setUrl] = useState("");
  const [name, setName] = useState("");
  const [groupId, setGroupId] = useState<string>("");
  const [proxy, setProxy] = useState("");
  const [isAdvancedOpen, setIsAdvancedOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isValidating, setIsValidating] = useState(false);
  const [detectedFeeds, setDetectedFeeds] = useState<DiscoveredFeed[]>([]);
  const [isFeedSelectOpen, setIsFeedSelectOpen] = useState(false);

  const resetForm = () => {
    setUrl("");
    setName("");
    setGroupId("");
    setProxy("");
    setIsAdvancedOpen(false);
    setDetectedFeeds([]);
    setIsFeedSelectOpen(false);
  };

  const handleClose = () => {
    setAddFeedOpen(false);
    resetForm();
  };

  const handleSelectDetectedFeed = (feed: DiscoveredFeed) => {
    setUrl(feed.link);
    setName((prev) => {
      if (prev.trim() || !feed.title.trim()) {
        return prev;
      }
      return feed.title.trim();
    });
    setIsFeedSelectOpen(false);
    setDetectedFeeds([]);
    toast.success(t("feed.toast.detected"));
  };

  const handleValidate = async () => {
    if (!url.trim()) return;

    setIsValidating(true);
    try {
      const response = await feedAPI.validate({ url: url.trim() });
      const feeds = response.data?.feeds ?? [];

      if (feeds.length === 0) {
        toast.info(t("feed.toast.noFeedsForUrl"));
        return;
      }

      if (feeds.length === 1) {
        handleSelectDetectedFeed(feeds[0]);
        return;
      }

      setDetectedFeeds(feeds);
      setIsFeedSelectOpen(true);
    } catch {
      toast.error(t("feed.toast.detectFailed"));
    } finally {
      setIsValidating(false);
    }
  };

  const handleSubmit = async () => {
    if (!url.trim()) {
      toast.error(t("feed.toast.enterUrl"));
      return;
    }

    const selectedGroupId = groupId
      ? parseInt(groupId, 10)
      : (groups[0]?.id ?? 1);

    setIsSubmitting(true);
    try {
      const request: CreateFeedRequest = {
        link: url.trim(),
        name: name.trim() || url.trim(),
        group_id: selectedGroupId,
      };

      if (proxy.trim()) {
        request.proxy = proxy.trim();
      }

      await createFeed.mutateAsync(request);
      toast.success(t("feed.toast.added"));
      handleClose();
    } catch {
      toast.error(t("feed.toast.addFailed"));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <>
      <Dialog open={isAddFeedOpen} onOpenChange={setAddFeedOpen}>
        <DialogContent
          className="flex w-full max-w-[480px] flex-col gap-0 overflow-hidden p-0"
          showCloseButton={false}
        >
          {/* Header */}
          <DialogHeader className="flex flex-row items-center justify-between border-b px-5 py-4">
            <DialogTitle className="text-base font-semibold">
              {t("feed.add.title")}
            </DialogTitle>
            <Button variant="ghost" size="icon-sm" onClick={handleClose}>
              <X className="h-[18px] w-[18px] text-muted-foreground" />
            </Button>
          </DialogHeader>

          {/* Form Content */}
          <div className="space-y-4 p-5">
            {/* URL Section */}
            <div className="space-y-1.5">
              <label className="text-[13px] font-medium">
                {t("feed.add.urlLabel")}
              </label>
              <div className="flex gap-2">
                <Input
                  placeholder={t("feed.add.urlPlaceholder")}
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  className="h-10"
                />
                <Button
                  variant="outline"
                  size="icon"
                  className="h-10 w-10 shrink-0"
                  onClick={handleValidate}
                  disabled={isValidating || !url.trim()}
                  title={t("feed.add.validateTitle")}
                >
                  <Radar
                    className={cn(
                      "h-[18px] w-[18px]",
                      isValidating && "animate-pulse",
                    )}
                  />
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">
                {t("feed.add.detectHint")}
              </p>
            </div>

            {/* Name Section */}
            <div className="space-y-1.5">
              <label className="text-[13px] font-medium">
                {t("feed.add.nameLabel")}
              </label>
              <Input
                placeholder={t("feed.add.namePlaceholder")}
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="h-10"
              />
            </div>

            {/* Group Section */}
            <div className="space-y-1.5">
              <label className="text-[13px] font-medium">
                {t("feed.add.groupLabel")}
              </label>
              <Select value={groupId} onValueChange={setGroupId}>
                <SelectTrigger className="h-10">
                  <SelectValue placeholder={t("feed.add.groupPlaceholder")} />
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

            {/* Advanced Section */}
            <Collapsible open={isAdvancedOpen} onOpenChange={setIsAdvancedOpen}>
              <CollapsibleTrigger className="flex w-full items-center gap-1.5 text-[13px] font-medium text-muted-foreground">
                <ChevronDown
                  className={cn(
                    "h-3.5 w-3.5 transition-transform",
                    isAdvancedOpen && "rotate-180",
                  )}
                />
                {t("feed.add.advanced")}
              </CollapsibleTrigger>
              <CollapsibleContent className="space-y-1.5 pl-5 pt-3">
                <label className="text-[13px] font-medium">
                  {t("feed.add.proxyLabel")}
                </label>
                <Input
                  placeholder={t("feed.add.proxyPlaceholder")}
                  value={proxy}
                  onChange={(e) => setProxy(e.target.value)}
                  className="h-10"
                />
                <p className="text-xs text-muted-foreground">
                  {t("feed.add.proxyHint")}
                </p>
              </CollapsibleContent>
            </Collapsible>
          </div>

          {/* Footer */}
          <div className="flex items-center justify-end gap-3 border-t px-5 py-4">
            <Button variant="outline" onClick={handleClose}>
              {t("common.cancel")}
            </Button>
            <Button
              onClick={handleSubmit}
              disabled={isSubmitting || !url.trim()}
            >
              <Plus className="mr-1.5 h-4 w-4" />
              {t("feed.add.button")}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog
        open={isFeedSelectOpen}
        onOpenChange={(open) => {
          setIsFeedSelectOpen(open);
          if (!open) {
            setDetectedFeeds([]);
          }
        }}
      >
        <DialogContent
          className="w-full max-w-[560px] p-0"
          showCloseButton={false}
        >
          <DialogHeader className="flex flex-row items-center justify-between border-b px-5 py-4">
            <div>
              <DialogTitle className="text-base font-semibold">
                {t("feed.select.title")}
              </DialogTitle>
              <DialogDescription>
                {t("feed.select.description")}
              </DialogDescription>
            </div>
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={() => {
                setIsFeedSelectOpen(false);
                setDetectedFeeds([]);
              }}
            >
              <X className="h-[18px] w-[18px] text-muted-foreground" />
            </Button>
          </DialogHeader>

          <div className="max-h-[360px] space-y-2 overflow-y-auto p-4">
            {detectedFeeds.map((feed, index) => (
              <button
                key={`${feed.link}-${index}`}
                type="button"
                onClick={() => handleSelectDetectedFeed(feed)}
                className="w-full rounded-md border p-3 text-left transition-colors hover:bg-accent/50"
              >
                <p className="truncate text-sm font-medium">
                  {feed.title || t("feed.select.fallback", { index: index + 1 })}
                </p>
                <p className="mt-1 truncate text-xs text-muted-foreground">
                  {feed.link}
                </p>
              </button>
            ))}
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
