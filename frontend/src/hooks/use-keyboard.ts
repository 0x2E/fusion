import { useEffect } from "react";
import { useUIStore } from "@/store";

export function useKeyboardShortcuts() {
  const { setSearchOpen, setSettingsOpen, isSearchOpen, isSettingsOpen, selectedArticleId, setSelectedArticle } =
    useUIStore();

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      // ⌘K or Ctrl+K: Open search
      if ((event.metaKey || event.ctrlKey) && event.key === "k") {
        event.preventDefault();
        setSearchOpen(!isSearchOpen);
        return;
      }

      // ESC: Close modals/drawer
      if (event.key === "Escape") {
        if (isSearchOpen) {
          setSearchOpen(false);
          return;
        }
        if (isSettingsOpen) {
          setSettingsOpen(false);
          return;
        }
        if (selectedArticleId !== null) {
          setSelectedArticle(null);
          return;
        }
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [isSearchOpen, isSettingsOpen, selectedArticleId, setSearchOpen, setSettingsOpen, setSelectedArticle]);
}

export function useArticleNavigation(articleIds: number[]) {
  const { selectedArticleId, setSelectedArticle } = useUIStore();

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      // Don't handle if we're in an input or if no article is selected
      if (
        event.target instanceof HTMLInputElement ||
        event.target instanceof HTMLTextAreaElement
      ) {
        return;
      }

      const currentIndex = selectedArticleId
        ? articleIds.indexOf(selectedArticleId)
        : -1;

      // J or ArrowDown: Next article
      if (event.key === "j" || event.key === "ArrowDown") {
        event.preventDefault();
        if (currentIndex < articleIds.length - 1) {
          setSelectedArticle(articleIds[currentIndex + 1]);
        }
        return;
      }

      // K or ArrowUp: Previous article
      if (event.key === "k" || event.key === "ArrowUp") {
        event.preventDefault();
        if (currentIndex > 0) {
          setSelectedArticle(articleIds[currentIndex - 1]);
        }
        return;
      }

      // Enter: Open selected article
      if (event.key === "Enter" && currentIndex >= 0) {
        event.preventDefault();
        // Article is already selected, just keeping selection
        return;
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [articleIds, selectedArticleId, setSelectedArticle]);

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
