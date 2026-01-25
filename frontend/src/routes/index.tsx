import { createFileRoute } from "@tanstack/react-router";
import { AppLayout } from "@/components/layout/app-layout";
import { ArticleList } from "@/components/article/article-list";

export const Route = createFileRoute("/")({
  component: HomePage,
});

function HomePage() {
  return (
    <AppLayout>
      <ArticleList />
    </AppLayout>
  );
}
