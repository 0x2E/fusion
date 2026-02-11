import { ArticleList } from "@/components/article/article-list";
import { AppLayout } from "@/components/layout/app-layout";

export function ArticlePage() {
  return (
    <AppLayout>
      <ArticleList />
    </AppLayout>
  );
}
