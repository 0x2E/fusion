import { Menu } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetTrigger,
  SheetTitle,
} from "@/components/ui/sheet";
import { Sidebar } from "./sidebar";
import { ArticleDrawer } from "@/components/article/article-drawer";
import { SearchDialog } from "@/components/search/search-dialog";
import { SettingsDialog } from "@/components/settings/settings-dialog";
import { AddGroupDialog } from "@/components/group/add-group-dialog";
import { AddFeedDialog } from "@/components/feed/add-feed-dialog";
import { EditFeedDialog } from "@/components/feed/edit-feed-dialog";
import { ImportOpmlDialog } from "@/components/feed/import-opml-dialog";
import { useKeyboardShortcuts } from "@/hooks/use-keyboard";
import { useIsMobile } from "@/hooks/use-mobile";

interface AppLayoutProps {
  children: React.ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
  const isMobile = useIsMobile();

  // Register global keyboard shortcuts
  useKeyboardShortcuts();

  return (
    <div className="flex h-screen w-full overflow-hidden">
      {/* Desktop sidebar */}
      {!isMobile && <Sidebar />}

      {/* Mobile sidebar */}
      {isMobile && (
        <Sheet>
          <SheetTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="fixed left-2 top-2 z-40 md:hidden"
            >
              <Menu className="h-5 w-5" />
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-[260px] p-0">
            <SheetTitle className="sr-only">Navigation</SheetTitle>
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
    </div>
  );
}
