// The `feeds_` prefix keeps this a top-level route instead of nesting it under
// the `/feeds` management page route (feeds.lazy.tsx). The trailing underscore
// drops the segment from the URL, so the final path is /feeds/:feedId/:filter.
import { createLazyFileRoute } from "@tanstack/react-router";
import { ArticlePage } from "@/components/article/article-page";

export const Route = createLazyFileRoute("/feeds_/$feedId/$filter")({
  component: ArticlePage,
});
