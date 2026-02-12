import { create } from "zustand";
import type { Feed } from "@/lib/api";

interface UIState {
  // Modals
  isSearchOpen: boolean;
  isSettingsOpen: boolean;
  isAddGroupOpen: boolean;
  isAddFeedOpen: boolean;
  isEditFeedOpen: boolean;
  editingFeed: Feed | null;
  isImportOpmlOpen: boolean;
  isShortcutsOpen: boolean;

  // Mobile sidebar
  isSidebarOpen: boolean;

  // Actions
  setSearchOpen: (open: boolean) => void;
  setSettingsOpen: (open: boolean) => void;
  setAddGroupOpen: (open: boolean) => void;
  setAddFeedOpen: (open: boolean) => void;
  setEditFeedOpen: (open: boolean, feed?: Feed) => void;
  setImportOpmlOpen: (open: boolean) => void;
  setShortcutsOpen: (open: boolean) => void;
  setSidebarOpen: (open: boolean) => void;
}

export const useUIStore = create<UIState>((set) => ({
  isSearchOpen: false,
  isSettingsOpen: false,
  isAddGroupOpen: false,
  isAddFeedOpen: false,
  isEditFeedOpen: false,
  editingFeed: null,
  isImportOpmlOpen: false,
  isShortcutsOpen: false,
  isSidebarOpen: false,

  setSearchOpen: (open) => set({ isSearchOpen: open }),

  setSettingsOpen: (open) => set({ isSettingsOpen: open }),

  setAddGroupOpen: (open) => set({ isAddGroupOpen: open }),

  setAddFeedOpen: (open) => set({ isAddFeedOpen: open }),

  setEditFeedOpen: (open, feed) =>
    set({ isEditFeedOpen: open, editingFeed: open ? (feed ?? null) : null }),

  setImportOpmlOpen: (open) => set({ isImportOpmlOpen: open }),

  setShortcutsOpen: (open) => set({ isShortcutsOpen: open }),

  setSidebarOpen: (open) => set({ isSidebarOpen: open }),
}));
