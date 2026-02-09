import { createLazyFileRoute } from "@tanstack/react-router";
import { ArticleList } from "@/components/article/article-list";
import { AppLayout } from "@/components/layout/app-layout";

export const Route = createLazyFileRoute("/")({
  component: HomePage,
});

function HomePage() {
  return (
    <AppLayout>
      <ArticleList />
    </AppLayout>
  );
}
