import { useEffect, useRef } from "react";
import { useNavigate } from "@tanstack/react-router";
import { useUIStore } from "@/store";
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
  const pendingPrefixRef = useRef<"g" | null>(null);
  const pendingPrefixTimerRef = useRef<number | null>(null);

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
      if (event.defaultPrevented) {
        return;
      }

      const key = event.key.toLowerCase();

      // ⌘K or Ctrl+K: Open search
      if ((event.metaKey || event.ctrlKey) && key === "k") {
        event.preventDefault();
        setSearchOpen(!isSearchOpen);
        resetPrefix();
        return;
      }

      // ⌘, or Ctrl+, : Open settings
      if ((event.metaKey || event.ctrlKey) && event.key === ",") {
        event.preventDefault();
        setSettingsOpen(true);
        resetPrefix();
        return;
      }

      // ESC: Close modals/drawer
      if (event.key === "Escape") {
        if (isSearchOpen) {
          resetPrefix();
          setSearchOpen(false);
          return;
        }
        if (isSettingsOpen) {
          resetPrefix();
          setSettingsOpen(false);
          return;
        }
        if (isShortcutsOpen) {
          resetPrefix();
          setShortcutsOpen(false);
          return;
        }
        if (selectedArticleId !== null) {
          resetPrefix();
          setSelectedArticle(null);
          return;
        }

        resetPrefix();
        return;
      }

      if (event.metaKey || event.ctrlKey || event.altKey) {
        return;
      }

      if (isTypingTarget(event.target)) {
        return;
      }

      if (isSearchOpen || isSettingsOpen || isShortcutsOpen) {
        return;
      }

      if (pendingPrefixRef.current === "g") {
        if (key === "u") {
          event.preventDefault();
          selectTopLevelFilter("unread");
          resetPrefix();
          return;
        }
        if (key === "a") {
          event.preventDefault();
          selectTopLevelFilter("all");
          resetPrefix();
          return;
        }
        if (key === "s") {
          event.preventDefault();
          selectTopLevelFilter("starred");
          resetPrefix();
          return;
        }
        if (key === "f") {
          event.preventDefault();
          navigate({ to: "/feeds" });
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
        setSearchOpen(true);
        return;
      }

      // ?: Open shortcuts help
      if (event.key === "?") {
        event.preventDefault();
        setShortcutsOpen(true);
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
  }, [
    isSearchOpen,
    isSettingsOpen,
    isShortcutsOpen,
    navigate,
    selectTopLevelFilter,
    selectedArticleId,
    setSearchOpen,
    setSettingsOpen,
    setShortcutsOpen,
    setSelectedArticle,
  ]);
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

  useEffect(() => {
    if (!enabled) {
      return;
    }

    function handleKeyDown(event: KeyboardEvent) {
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

      if (
        isSearchOpen ||
        isSettingsOpen ||
        isAddGroupOpen ||
        isAddFeedOpen ||
        isEditFeedOpen ||
        isImportOpmlOpen ||
        isShortcutsOpen
      ) {
        return;
      }

      const currentIndex = selectedArticleId
        ? articleIds.indexOf(selectedArticleId)
        : -1;

      const key = event.key.toLowerCase();

      // J or ArrowDown: Next article
      if (key === "j" || key === "n" || event.key === "ArrowDown") {
        event.preventDefault();
        if (currentIndex < articleIds.length - 1) {
          setSelectedArticle(articleIds[currentIndex + 1]);
        }
        return;
      }

      // K or ArrowUp: Previous article
      if (key === "k" || key === "p" || event.key === "ArrowUp") {
        event.preventDefault();
        if (currentIndex > 0) {
          setSelectedArticle(articleIds[currentIndex - 1]);
        }
        return;
      }

      if (selectedArticleId === null) {
        return;
      }

      // M: toggle read/unread
      if (key === "m" && onToggleRead) {
        event.preventDefault();
        void onToggleRead();
        return;
      }

      // S/F: toggle star
      if ((key === "s" || key === "f") && onToggleStar) {
        event.preventDefault();
        void onToggleStar();
        return;
      }

      // O/V: open original article
      if ((key === "o" || key === "v") && onOpenOriginal) {
        event.preventDefault();
        onOpenOriginal();
        return;
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [
    articleIds,
    enabled,
    isAddFeedOpen,
    isAddGroupOpen,
    isEditFeedOpen,
    isImportOpmlOpen,
    isSearchOpen,
    isSettingsOpen,
    isShortcutsOpen,
    onOpenOriginal,
    onToggleRead,
    onToggleStar,
    selectedArticleId,
    setSelectedArticle,
  ]);

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
