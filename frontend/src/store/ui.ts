import { create } from "zustand";
import type { Feed } from "@/lib/api";

interface UIState {
  // Modals
  isSearchOpen: boolean;
  isSettingsOpen: boolean;
  isGroupManagementOpen: boolean;
  isAddFeedOpen: boolean;
  isEditFeedOpen: boolean;
  editingFeed: Feed | null;
  isFeedManagementOpen: boolean;
  isImportOpmlOpen: boolean;

  // Actions
  setSearchOpen: (open: boolean) => void;
  setSettingsOpen: (open: boolean) => void;
  setGroupManagementOpen: (open: boolean) => void;
  setAddFeedOpen: (open: boolean) => void;
  setEditFeedOpen: (open: boolean, feed?: Feed) => void;
  setFeedManagementOpen: (open: boolean) => void;
  setImportOpmlOpen: (open: boolean) => void;
}

export const useUIStore = create<UIState>((set) => ({
  isSearchOpen: false,
  isSettingsOpen: false,
  isGroupManagementOpen: false,
  isAddFeedOpen: false,
  isEditFeedOpen: false,
  editingFeed: null,
  isFeedManagementOpen: false,
  isImportOpmlOpen: false,

  setSearchOpen: (open) => set({ isSearchOpen: open }),

  setSettingsOpen: (open) => set({ isSettingsOpen: open }),

  setGroupManagementOpen: (open) => set({ isGroupManagementOpen: open }),

  setAddFeedOpen: (open) => set({ isAddFeedOpen: open }),

  setEditFeedOpen: (open, feed) =>
    set({ isEditFeedOpen: open, editingFeed: open ? feed ?? null : null }),

  setFeedManagementOpen: (open) => set({ isFeedManagementOpen: open }),

  setImportOpmlOpen: (open) => set({ isImportOpmlOpen: open }),
}));
