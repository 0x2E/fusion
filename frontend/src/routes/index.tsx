import { createFileRoute, redirect } from "@tanstack/react-router";
import { defaultArticleFilter } from "@/lib/article-filter";

export const Route = createFileRoute("/")({
  beforeLoad: () => {
    throw redirect({
      to: "/$filter",
      params: { filter: defaultArticleFilter },
      replace: true,
    });
  },
});
