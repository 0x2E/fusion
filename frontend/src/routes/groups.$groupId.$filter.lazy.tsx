import { createLazyFileRoute } from "@tanstack/react-router";
import { ArticlePage } from "@/components/article/article-page";

export const Route = createLazyFileRoute("/groups/$groupId/$filter")({
  component: ArticlePage,
});
