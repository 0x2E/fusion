import { RSSReaderApp } from "@/components/rss-reader-app";
import { createFileRoute } from "@tanstack/react-router";

interface SearchParams {
  feed?: number;
  group?: number;
  filter?: "all" | "unread" | "starred";
  item?: number;
  search?: boolean;
  settings?: boolean;
}

export const Route = createFileRoute("/")({
  component: RSSReaderApp,
  validateSearch: (search: Record<string, unknown>): SearchParams => {
    return {
      feed: typeof search.feed === "number" ? search.feed : undefined,
      group: typeof search.group === "number" ? search.group : undefined,
      filter:
        search.filter === "unread" || search.filter === "starred"
          ? search.filter
          : "all",
      item: typeof search.item === "number" ? search.item : undefined,
      search: search.search === true,
      settings: search.settings === true,
    };
  },
});
