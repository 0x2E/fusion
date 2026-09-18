import { useEffect, useRef } from "react";
import { useNavigate } from "@tanstack/react-router";
import {
  sidebarNodeKey,
  useSidebarNavStore,
  useUIStore,
  type SidebarNode,
} from "@/store";
import { useIsMobile } from "@/hooks/use-mobile";
import { useUrlState } from "./use-url-state";

function isTypingTarget(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) {
    return false;
  }

  if (target.isContentEditable) {
    return true;
  }

  const tagName = target.tagName;
  if (tagName === "INPUT" || tagName === "TEXTAREA" || tagName === "SELECT") {
    return true;
  }

  return target.closest("[contenteditable='true'], [data-hotkey-ignore='true']") !== null;
}

interface ArticleNavigationOptions {
  enabled?: boolean;
  onToggleRead?: () => void | Promise<void>;
  onToggleStar?: () => void | Promise<void>;
  onOpenOriginal?: () => void;
}

export function useKeyboardShortcuts() {
  const setSearchOpen = useUIStore((s) => s.setSearchOpen);
  const setSettingsOpen = useUIStore((s) => s.setSettingsOpen);
  const setShortcutsOpen = useUIStore((s) => s.setShortcutsOpen);
  const isSearchOpen = useUIStore((s) => s.isSearchOpen);
  const isSettingsOpen = useUIStore((s) => s.isSettingsOpen);
  const isShortcutsOpen = useUIStore((s) => s.isShortcutsOpen);
  const { selectedArticleId, setSelectedArticle, selectTopLevelFilter } =
    useUrlState();
  const navigate = useNavigate();
  const setSidebarFocusedKey = useSidebarNavStore((s) => s.setFocusedKey);
  const pendingPrefixRef = useRef<"g" | null>(null);
  const pendingPrefixTimerRef = useRef<number | null>(null);
  const latestStateRef = useRef({
    isSearchOpen,
    isSettingsOpen,
    isShortcutsOpen,
    selectedArticleId,
    setSearchOpen,
    setSettingsOpen,
    setShortcutsOpen,
    setSelectedArticle,
    selectTopLevelFilter,
    setSidebarFocusedKey,
    navigate,
  });

  useEffect(() => {
    latestStateRef.current = {
      isSearchOpen,
      isSettingsOpen,
      isShortcutsOpen,
      selectedArticleId,
      setSearchOpen,
      setSettingsOpen,
      setShortcutsOpen,
      setSelectedArticle,
      selectTopLevelFilter,
      setSidebarFocusedKey,
      navigate,
    };
  }, [
    isSearchOpen,
    isSettingsOpen,
    isShortcutsOpen,
    selectedArticleId,
    setSearchOpen,
    setSettingsOpen,
    setShortcutsOpen,
    setSelectedArticle,
    selectTopLevelFilter,
    setSidebarFocusedKey,
    navigate,
  ]);

  useEffect(() => {
    function resetPrefix() {
      pendingPrefixRef.current = null;
      if (pendingPrefixTimerRef.current !== null) {
        window.clearTimeout(pendingPrefixTimerRef.current);
        pendingPrefixTimerRef.current = null;
      }
    }

    function startPrefix(prefix: "g") {
      resetPrefix();
      pendingPrefixRef.current = prefix;
      pendingPrefixTimerRef.current = window.setTimeout(() => {
        resetPrefix();
      }, 1200);
    }

    function handleKeyDown(event: KeyboardEvent) {
      const state = latestStateRef.current;

      if (event.defaultPrevented) {
        return;
      }

      const key = event.key.toLowerCase();

      // ⌘K or Ctrl+K: Open search
      if ((event.metaKey || event.ctrlKey) && key === "k") {
        event.preventDefault();
        state.setSearchOpen(!state.isSearchOpen);
        resetPrefix();
        return;
      }

      // ⌘, or Ctrl+, : Open settings
      if ((event.metaKey || event.ctrlKey) && event.key === ",") {
        event.preventDefault();
        state.setSettingsOpen(true);
        resetPrefix();
        return;
      }

      // ESC: Close modals/drawer
      if (event.key === "Escape") {
        if (state.isSearchOpen) {
          resetPrefix();
          state.setSearchOpen(false);
          return;
        }
        if (state.isSettingsOpen) {
          resetPrefix();
          state.setSettingsOpen(false);
          return;
        }
        if (state.isShortcutsOpen) {
          resetPrefix();
          state.setShortcutsOpen(false);
          return;
        }
        if (state.selectedArticleId !== null) {
          resetPrefix();
          state.setSelectedArticle(null);
          return;
        }

        // Clear the sidebar highlight when nothing else is open.
        state.setSidebarFocusedKey(null);
        resetPrefix();
        return;
      }

      if (event.metaKey || event.ctrlKey || event.altKey) {
        return;
      }

      if (isTypingTarget(event.target)) {
        return;
      }

      if (state.isSearchOpen || state.isSettingsOpen || state.isShortcutsOpen) {
        return;
      }

      if (pendingPrefixRef.current === "g") {
        if (key === "u") {
          event.preventDefault();
          state.selectTopLevelFilter("unread");
          resetPrefix();
          return;
        }
        if (key === "a") {
          event.preventDefault();
          state.selectTopLevelFilter("all");
          resetPrefix();
          return;
        }
        if (key === "s") {
          event.preventDefault();
          state.selectTopLevelFilter("starred");
          resetPrefix();
          return;
        }
        if (key === "f") {
          event.preventDefault();
          state.navigate({ to: "/feeds" });
          resetPrefix();
          return;
        }

        resetPrefix();
      }

      if (key === "g") {
        event.preventDefault();
        startPrefix("g");
        return;
      }

      // /: Open search
      if (event.key === "/") {
        event.preventDefault();
        state.setSearchOpen(true);
        return;
      }

      // ? or Shift+H: Open shortcuts help
      if (event.key === "?" || (event.shiftKey && key === "h")) {
        event.preventDefault();
        state.setShortcutsOpen(true);
        return;
      }
    }

    document.addEventListener("keydown", handleKeyDown, true);
    return () => {
      if (pendingPrefixTimerRef.current !== null) {
        window.clearTimeout(pendingPrefixTimerRef.current);
      }
      document.removeEventListener("keydown", handleKeyDown, true);
    };
  }, []);
}

export function useArticleNavigation(
  articleIds: number[],
  options: ArticleNavigationOptions = {},
) {
  const { selectedArticleId, setSelectedArticle } = useUrlState();
  const isSearchOpen = useUIStore((s) => s.isSearchOpen);
  const isSettingsOpen = useUIStore((s) => s.isSettingsOpen);
  const isAddGroupOpen = useUIStore((s) => s.isAddGroupOpen);
  const isAddFeedOpen = useUIStore((s) => s.isAddFeedOpen);
  const isEditFeedOpen = useUIStore((s) => s.isEditFeedOpen);
  const isImportOpmlOpen = useUIStore((s) => s.isImportOpmlOpen);
  const isShortcutsOpen = useUIStore((s) => s.isShortcutsOpen);
  const {
    enabled = true,
    onToggleRead,
    onToggleStar,
    onOpenOriginal,
  } = options;
  const latestStateRef = useRef({
    articleIds,
    enabled,
    selectedArticleId,
    isSearchOpen,
    isSettingsOpen,
    isAddGroupOpen,
    isAddFeedOpen,
    isEditFeedOpen,
    isImportOpmlOpen,
    isShortcutsOpen,
    onToggleRead,
    onToggleStar,
    onOpenOriginal,
    setSelectedArticle,
  });

  useEffect(() => {
    latestStateRef.current = {
      articleIds,
      enabled,
      selectedArticleId,
      isSearchOpen,
      isSettingsOpen,
      isAddGroupOpen,
      isAddFeedOpen,
      isEditFeedOpen,
      isImportOpmlOpen,
      isShortcutsOpen,
      onToggleRead,
      onToggleStar,
      onOpenOriginal,
      setSelectedArticle,
    };
  }, [
    articleIds,
    enabled,
    selectedArticleId,
    isSearchOpen,
    isSettingsOpen,
    isAddGroupOpen,
    isAddFeedOpen,
    isEditFeedOpen,
    isImportOpmlOpen,
    isShortcutsOpen,
    onToggleRead,
    onToggleStar,
    onOpenOriginal,
    setSelectedArticle,
  ]);

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      const state = latestStateRef.current;

      if (!state.enabled) {
        return;
      }

      if (event.defaultPrevented) {
        return;
      }

      // Don't handle navigation keys while typing in form fields
      if (isTypingTarget(event.target)) {
        return;
      }

      if (event.metaKey || event.ctrlKey || event.altKey) {
        return;
      }

      // Shift chords are reserved for sidebar tree navigation.
      if (event.shiftKey) {
        return;
      }

      if (
        state.isSearchOpen ||
        state.isSettingsOpen ||
        state.isAddGroupOpen ||
        state.isAddFeedOpen ||
        state.isEditFeedOpen ||
        state.isImportOpmlOpen ||
        state.isShortcutsOpen
      ) {
        return;
      }

      const currentIndex = state.selectedArticleId
        ? state.articleIds.indexOf(state.selectedArticleId)
        : -1;

      const key = event.key.toLowerCase();

      // J or ArrowDown: Next article
      if (key === "j" || key === "n" || event.key === "ArrowDown") {
        event.preventDefault();
        if (currentIndex < state.articleIds.length - 1) {
          state.setSelectedArticle(state.articleIds[currentIndex + 1]);
        }
        return;
      }

      // K or ArrowUp: Previous article
      if (key === "k" || key === "p" || event.key === "ArrowUp") {
        event.preventDefault();
        if (currentIndex > 0) {
          state.setSelectedArticle(state.articleIds[currentIndex - 1]);
        }
        return;
      }

      if (state.selectedArticleId === null) {
        return;
      }

      // M: toggle read/unread
      if (key === "m" && state.onToggleRead) {
        event.preventDefault();
        void state.onToggleRead();
        return;
      }

      // S/F: toggle star
      if ((key === "s" || key === "f") && state.onToggleStar) {
        event.preventDefault();
        void state.onToggleStar();
        return;
      }

      // O/V: open original article
      if ((key === "o" || key === "v") && state.onOpenOriginal) {
        event.preventDefault();
        state.onOpenOriginal();
        return;
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);

  const goToNext = () => {
    const currentIndex = selectedArticleId
      ? articleIds.indexOf(selectedArticleId)
      : -1;
    if (currentIndex < articleIds.length - 1) {
      setSelectedArticle(articleIds[currentIndex + 1]);
    }
  };

  const goToPrevious = () => {
    const currentIndex = selectedArticleId
      ? articleIds.indexOf(selectedArticleId)
      : -1;
    if (currentIndex > 0) {
      setSelectedArticle(articleIds[currentIndex - 1]);
    }
  };

  const hasNext = () => {
    const currentIndex = selectedArticleId
      ? articleIds.indexOf(selectedArticleId)
      : -1;
    return currentIndex < articleIds.length - 1;
  };

  const hasPrevious = () => {
    const currentIndex = selectedArticleId
      ? articleIds.indexOf(selectedArticleId)
      : -1;
    return currentIndex > 0;
  };

  return { goToNext, goToPrevious, hasNext, hasPrevious };
}

/**
 * Inoreader-style sidebar tree navigation:
 * Shift+N / Shift+P move the highlight, Shift+J / Shift+K move and open the
 * item, Shift+O opens the focused item and Shift+X expands or collapses its
 * group.
 */
export function useSidebarNavigation(nodes: SidebarNode[]) {
  const { selectTopLevelFilter, setSelectedGroup, setSelectedFeed } =
    useUrlState();
  const isMobile = useIsMobile();
  const isSidebarOpen = useUIStore((s) => s.isSidebarOpen);
  const setFocusedKey = useSidebarNavStore((s) => s.setFocusedKey);
  const focusedKey = useSidebarNavStore((s) => s.focusedKey);
  const collapsedGroupIds = useSidebarNavStore((s) => s.collapsedGroupIds);
  const setGroupCollapsed = useSidebarNavStore((s) => s.setGroupCollapsed);
  const isSearchOpen = useUIStore((s) => s.isSearchOpen);
  const isSettingsOpen = useUIStore((s) => s.isSettingsOpen);
  const isAddGroupOpen = useUIStore((s) => s.isAddGroupOpen);
  const isAddFeedOpen = useUIStore((s) => s.isAddFeedOpen);
  const isEditFeedOpen = useUIStore((s) => s.isEditFeedOpen);
  const isImportOpmlOpen = useUIStore((s) => s.isImportOpmlOpen);
  const isShortcutsOpen = useUIStore((s) => s.isShortcutsOpen);

  const latestStateRef = useRef({
    nodes,
    focusedKey,
    collapsedGroupIds,
    isMobile,
    isSidebarOpen,
    selectTopLevelFilter,
    setSelectedGroup,
    setSelectedFeed,
    setFocusedKey,
    setGroupCollapsed,
    isSearchOpen,
    isSettingsOpen,
    isAddGroupOpen,
    isAddFeedOpen,
    isEditFeedOpen,
    isImportOpmlOpen,
    isShortcutsOpen,
  });

  useEffect(() => {
    latestStateRef.current = {
      nodes,
      focusedKey,
      collapsedGroupIds,
      isMobile,
      isSidebarOpen,
      selectTopLevelFilter,
      setSelectedGroup,
      setSelectedFeed,
      setFocusedKey,
      setGroupCollapsed,
      isSearchOpen,
      isSettingsOpen,
      isAddGroupOpen,
      isAddFeedOpen,
      isEditFeedOpen,
      isImportOpmlOpen,
      isShortcutsOpen,
    };
  }, [
    nodes,
    focusedKey,
    collapsedGroupIds,
    isMobile,
    isSidebarOpen,
    selectTopLevelFilter,
    setSelectedGroup,
    setSelectedFeed,
    setFocusedKey,
    setGroupCollapsed,
    isSearchOpen,
    isSettingsOpen,
    isAddGroupOpen,
    isAddFeedOpen,
    isEditFeedOpen,
    isImportOpmlOpen,
    isShortcutsOpen,
  ]);

  // Keep the highlighted item visible within the sidebar scroll area.
  useEffect(() => {
    if (focusedKey === null) {
      return;
    }
    document
      .querySelector(`[data-sidebar-key="${focusedKey}"]`)
      ?.scrollIntoView({ block: "nearest" });
  }, [focusedKey]);

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (
        !event.shiftKey ||
        event.metaKey ||
        event.ctrlKey ||
        event.altKey
      ) {
        return;
      }

      if (event.defaultPrevented || isTypingTarget(event.target)) {
        return;
      }

      const state = latestStateRef.current;

      if (
        state.isSearchOpen ||
        state.isSettingsOpen ||
        state.isAddGroupOpen ||
        state.isAddFeedOpen ||
        state.isEditFeedOpen ||
        state.isImportOpmlOpen ||
        state.isShortcutsOpen
      ) {
        return;
      }

      // On mobile the sidebar lives in a Sheet; ignore chords while it is hidden.
      if (state.isMobile && !state.isSidebarOpen) {
        return;
      }

      const key = event.key.toLowerCase();

      const activate = (node: SidebarNode) => {
        if (node.kind === "filter") {
          state.selectTopLevelFilter(node.filter);
        } else if (node.kind === "group") {
          state.setSelectedGroup(node.groupId);
        } else {
          state.setSelectedFeed(node.feedId);
        }
      };

      if (key === "n" || key === "j" || key === "p" || key === "k") {
        if (state.nodes.length === 0) {
          return;
        }
        event.preventDefault();

        const currentIndex = state.nodes.findIndex(
          (node) => node.key === state.focusedKey,
        );
        const direction = key === "n" || key === "j" ? 1 : -1;
        const nextIndex =
          currentIndex === -1
            ? direction === 1
              ? 0
              : state.nodes.length - 1
            : Math.min(
                Math.max(currentIndex + direction, 0),
                state.nodes.length - 1,
              );

        const nextNode = state.nodes[nextIndex];
        state.setFocusedKey(nextNode.key);

        // Shift+J / Shift+K move and open; Shift+N / Shift+P only move.
        if (key === "j" || key === "k") {
          activate(nextNode);
        }
        return;
      }

      const node = state.nodes.find((item) => item.key === state.focusedKey);
      if (!node) {
        return;
      }

      if (key === "o") {
        event.preventDefault();
        activate(node);
        return;
      }

      if (key === "x") {
        event.preventDefault();
        const groupId =
          node.kind === "group"
            ? node.groupId
            : node.kind === "feed"
              ? node.groupId
              : null;
        // Ungrouped feeds have no folder to collapse.
        if (groupId === null || groupId === 0) {
          return;
        }
        const willCollapse = !state.collapsedGroupIds.includes(groupId);
        state.setGroupCollapsed(groupId, willCollapse);

        // Collapsing the parent hides the focused feed; keep focus on the group.
        if (willCollapse && node.kind === "feed") {
          state.setFocusedKey(sidebarNodeKey.group(groupId));
        }
      }
    }

    document.addEventListener("keydown", handleKeyDown, true);
    return () => document.removeEventListener("keydown", handleKeyDown, true);
  }, []);
}
