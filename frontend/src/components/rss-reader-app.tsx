import { useDataStore } from "@/store/data";
import { useUIStore } from "@/store/ui";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { useEffect } from "react";
import { ArticleDrawer } from "./article-drawer";
import { ArticleList } from "./article-list";
import { MobileSidebar } from "./mobile-sidebar";
import { SearchDialog } from "./search-dialog";
import { SettingsDialog } from "./settings-dialog";
import { Sidebar } from "./sidebar";

export function RSSReaderApp() {
  const navigate = useNavigate();
  const search = useSearch({ from: "/" });
  const { fetchGroups, fetchFeeds } = useDataStore();
  const { sidebarOpen, setSidebarOpen } = useUIStore();

  useEffect(() => {
    fetchGroups();
    fetchFeeds();
  }, [fetchGroups, fetchFeeds]);

  const handleUpdateSearch = (updates: Partial<typeof search>) => {
    navigate({
      to: "/",
      search: { ...search, ...updates },
    });
  };

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      {/* Desktop Sidebar */}
      <aside className="hidden md:block">
        <Sidebar
          onSearchOpen={() => handleUpdateSearch({ search: true })}
          onSettingsOpen={() => handleUpdateSearch({ settings: true })}
        />
      </aside>

      {/* Mobile Sidebar */}
      <MobileSidebar
        open={sidebarOpen}
        onOpenChange={setSidebarOpen}
        onSearchOpen={() => {
          handleUpdateSearch({ search: true });
          setSidebarOpen(false);
        }}
        onSettingsOpen={() => {
          handleUpdateSearch({ settings: true });
          setSidebarOpen(false);
        }}
      />

      {/* Main Content */}
      <main className="flex-1 flex flex-col h-full overflow-hidden">
        <ArticleList onMenuOpen={() => setSidebarOpen(true)} />
      </main>

      {/* Article Drawer */}
      <ArticleDrawer />

      {/* Dialogs */}
      <SearchDialog
        open={search.search ?? false}
        onOpenChange={(open) => handleUpdateSearch({ search: open })}
      />
      <SettingsDialog
        open={search.settings ?? false}
        onOpenChange={(open) => handleUpdateSearch({ settings: open })}
      />
    </div>
  );
}
