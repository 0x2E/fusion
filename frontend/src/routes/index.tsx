import { createFileRoute } from "@tanstack/react-router";

export interface SearchParams {
  feed?: number;
  group?: number;
  filter?: "all" | "unread" | "starred";
  article?: number;
}

export const Route = createFileRoute("/")({
  validateSearch: (search: Record<string, unknown>): SearchParams => ({
    feed: typeof search.feed === "number" ? search.feed : undefined,
    group: typeof search.group === "number" ? search.group : undefined,
    filter:
      search.filter === "all" || search.filter === "unread" || search.filter === "starred"
        ? search.filter
        : undefined,
    article: typeof search.article === "number" ? search.article : undefined,
  }),
});
