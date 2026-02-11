import { createLazyFileRoute } from "@tanstack/react-router";
import { ArticlePage } from "@/components/article/article-page";

export const Route = createLazyFileRoute("/feeds_/$feedId/$filter")({
  component: ArticlePage,
});
