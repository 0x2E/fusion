import { CircleAlert, LogOut, Search, Settings, Rss } from "lucide-react";
import { useLocation, useNavigate } from "@tanstack/react-router";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { FeedList } from "@/components/feed/feed-list";
import {
  Popover,
  PopoverContent,
  PopoverTitle,
  PopoverTrigger,
} from "@/components/ui/popover";
import { sessionAPI } from "@/lib/api";
import { useI18n, type TranslationKey } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import { useAppInfo } from "@/queries/app";
import { useUIStore } from "@/store";
import { resetSessionGuard } from "@/routes/__root";

// Notice codes from GET /api/app. Unknown codes are ignored so the backend can
// ship new notices before the client knows their copy.
const noticeMessages: Record<
  string,
  { title: TranslationKey; body: TranslationKey }
> = {
  oidc_allowed_user_unset: {
    title: "sidebar.notices.oidcAllowedUserUnset.title",
    body: "sidebar.notices.oidcAllowedUserUnset.body",
  },
};

export function Sidebar() {
  const { t } = useI18n();
  const { setSearchOpen, setSettingsOpen } = useUIStore();
  const { data: appInfo } = useAppInfo();
  const notices = (appInfo?.notices ?? []).flatMap((code) => {
    const message = noticeMessages[code];
    return message ? [message] : [];
  });
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { pathname } = useLocation();
  const isFeedsPage = pathname === "/feeds";

  const handleSignOut = async () => {
    if (!window.confirm(t("sidebar.signOutConfirm"))) {
      return;
    }
    try {
      await sessionAPI.logout();
    } catch {
      toast.error(t("sidebar.signOutFailed"));
      return;
    }

    queryClient.clear();
    // A full-page navigation leaves this document so the route guard starts
    // fresh; the SPA router would keep its cached "session verified" verdict.
    resetSessionGuard();
    window.location.assign("/login");
  };

  return (
    <aside className="sidebar-typography flex h-full w-75 flex-none flex-col overflow-hidden border-r bg-sidebar text-sidebar-foreground">
      {/* Header */}
      <div className="flex h-14 shrink-0 items-center gap-2 border-b px-4">
        <img
          src="/icon-96.png"
          alt={t("common.fusionLogo")}
          width={32}
          height={32}
          className="h-8 w-8 rounded-md"
        />
        <span className="text-base font-semibold">Fusion</span>
        {notices.length > 0 && (
          <Popover>
            <PopoverTrigger
              aria-label={t("sidebar.notices.label")}
              className="ml-auto rounded-md p-1.5 text-amber-600 outline-none transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring/50"
            >
              <CircleAlert className="h-4 w-4" />
            </PopoverTrigger>
            <PopoverContent align="end" className="w-80 gap-3">
              {/* One title names the popover; each notice carries its own heading. */}
              <PopoverTitle className="sr-only">
                {t("sidebar.notices.label")}
              </PopoverTitle>
              {notices.map((notice) => (
                <div key={notice.title} className="flex flex-col gap-1">
                  <h3 className="font-medium">{t(notice.title)}</h3>
                  <p className="text-muted-foreground">{t(notice.body)}</p>
                </div>
              ))}
            </PopoverContent>
          </Popover>
        )}
      </div>

      {/* Search button */}
      <div className="px-2 pt-3">
        <button
          className="flex w-full items-center justify-between rounded-md bg-muted px-3 py-2 text-muted-foreground transition-colors hover:bg-accent"
          onClick={() => setSearchOpen(true)}
        >
          <div className="flex items-center gap-2">
            <Search className="h-4 w-4" />
            <span className="text-sm">{t("sidebar.search")}</span>
          </div>
          <kbd className="rounded bg-accent px-1.5 py-0.5 font-mono text-[11px] font-medium">
            Cmd+K / ?
          </kbd>
        </button>
      </div>

      {/* Feed list */}
      <FeedList />

      {/* Footer */}
      <div className="p-2">
        <button
          className={cn(
            "flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors",
            isFeedsPage
              ? "bg-accent text-accent-foreground"
              : "hover:bg-accent/50",
          )}
          onClick={() => navigate({ to: "/feeds" })}
        >
          <Rss className="h-4 w-4 shrink-0 text-muted-foreground" />
          <span>{t("sidebar.manageFeeds")}</span>
        </button>
        <button
          className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors hover:bg-accent/50"
          onClick={() => setSettingsOpen(true)}
        >
          <Settings className="h-4 w-4 shrink-0 text-muted-foreground" />
          <span>{t("sidebar.settings")}</span>
        </button>
        <button
          className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors hover:bg-accent/50"
          onClick={handleSignOut}
        >
          <LogOut className="h-4 w-4 shrink-0 text-muted-foreground" />
          <span>{t("sidebar.signOut")}</span>
        </button>
      </div>
    </aside>
  );
}
