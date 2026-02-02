import { createFileRoute } from "@tanstack/react-router";
import { AppLayout } from "@/components/layout/app-layout";
import { ArticleList } from "@/components/article/article-list";

export interface SearchParams {
  feed?: number;
  group?: number;
  filter?: "all" | "unread" | "starred";
  article?: number;
}

export const Route = createFileRoute("/")({
  component: HomePage,
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

function HomePage() {
  return (
    <AppLayout>
      <ArticleList />
    </AppLayout>
  );
}
