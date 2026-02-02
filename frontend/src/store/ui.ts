import { create } from "zustand";

interface UIState {
  // Modals
  isSearchOpen: boolean;
  isSettingsOpen: boolean;
  isGroupManagementOpen: boolean;
  isAddFeedOpen: boolean;
  isFeedManagementOpen: boolean;
  isImportOpmlOpen: boolean;

  // Actions
  setSearchOpen: (open: boolean) => void;
  setSettingsOpen: (open: boolean) => void;
  setGroupManagementOpen: (open: boolean) => void;
  setAddFeedOpen: (open: boolean) => void;
  setFeedManagementOpen: (open: boolean) => void;
  setImportOpmlOpen: (open: boolean) => void;
}

export const useUIStore = create<UIState>((set) => ({
  isSearchOpen: false,
  isSettingsOpen: false,
  isGroupManagementOpen: false,
  isAddFeedOpen: false,
  isFeedManagementOpen: false,
  isImportOpmlOpen: false,

  setSearchOpen: (open) => set({ isSearchOpen: open }),

  setSettingsOpen: (open) => set({ isSettingsOpen: open }),

  setGroupManagementOpen: (open) => set({ isGroupManagementOpen: open }),

  setAddFeedOpen: (open) => set({ isAddFeedOpen: open }),

  setFeedManagementOpen: (open) => set({ isFeedManagementOpen: open }),

  setImportOpmlOpen: (open) => set({ isImportOpmlOpen: open }),
}));
