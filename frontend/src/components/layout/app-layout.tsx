import { useEffect } from "react";
import { useLocation } from "@tanstack/react-router";
import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet";
import { Sidebar } from "./sidebar";
import { ArticleDrawer } from "@/components/article/article-drawer";
import { SearchDialog } from "@/components/search/search-dialog";
import { SettingsDialog } from "@/components/settings/settings-dialog";
import { AddGroupDialog } from "@/components/group/add-group-dialog";
import { AddFeedDialog } from "@/components/feed/add-feed-dialog";
import { EditFeedDialog } from "@/components/feed/edit-feed-dialog";
import { ImportOpmlDialog } from "@/components/feed/import-opml-dialog";
import { ShortcutsDialog } from "@/components/layout/shortcuts-dialog";
import { useKeyboardShortcuts } from "@/hooks/use-keyboard";
import { useI18n } from "@/lib/i18n";
import { useIsMobile } from "@/hooks/use-mobile";
import { useUIStore } from "@/store/ui";

interface AppLayoutProps {
  children: React.ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
  const { t } = useI18n();
  const isMobile = useIsMobile();
  const isSidebarOpen = useUIStore((s) => s.isSidebarOpen);
  const setSidebarOpen = useUIStore((s) => s.setSidebarOpen);

  // Register global keyboard shortcuts
  useKeyboardShortcuts();

  // Close mobile sidebar on navigation
  const location = useLocation();
  useEffect(() => {
    setSidebarOpen(false);
  }, [location.pathname, location.searchStr, setSidebarOpen]);

  return (
    <div className="flex h-screen w-full overflow-hidden">
      {/* Desktop sidebar */}
      {!isMobile && <Sidebar />}

      {/* Mobile sidebar */}
      {isMobile && (
        <Sheet open={isSidebarOpen} onOpenChange={setSidebarOpen}>
          <SheetContent
            side="left"
            className="w-[260px] p-0"
            showCloseButton={false}
          >
            <SheetTitle className="sr-only">{t("common.navigation")}</SheetTitle>
            <Sidebar />
          </SheetContent>
        </Sheet>
      )}

      {/* Main content */}
      <main className="flex-1 overflow-hidden">{children}</main>

      {/* Modals and drawers */}
      <ArticleDrawer />
      <SearchDialog />
      <SettingsDialog />
      <AddGroupDialog />
      <AddFeedDialog />
      <EditFeedDialog />
      <ImportOpmlDialog />
      <ShortcutsDialog />
    </div>
  );
}
